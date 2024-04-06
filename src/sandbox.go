package main

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

const DefaultSandboxExpiry = 30 * time.Minute
const DefaultInactivityExpirationThreshold = 20 * time.Minute
const DefaultCleanupFrequency = 5 * time.Minute

type UserSandbox struct {
	Id        int
	ExpiresAt time.Time
	LastUsed  time.Time
}

type Sandbox struct {
	mu              sync.Mutex
	Count           int
	Reserved        map[string]*UserSandbox
	AvailableBoxIDs map[int]struct{}
	SandboxExpiry   time.Duration
	// frequency to check sandbox expiry and inactivity
	CleanupFrequency time.Duration
	// if a sandbox assigned to session is inactive for this threshold,
	// 		it will be released from the user, cleaned up and made available for other sessions
	InactivityExpirationThreshold time.Duration
}

func NewSandbox(count int) *Sandbox {
	s := map[int]struct{}{}
	for i := 0; i < count; i++ {
		cmd := exec.Command("isolate", "--init", fmt.Sprintf("-b %v", i))
		op, err := cmd.Output()
		if err != nil {
			fmt.Printf("failed to create sandbox %v, output: %v err: %v", i, string(op), err)
		} else {
			s[i] = struct{}{}
		}
	}

	return &Sandbox{
		Count:                         count,
		AvailableBoxIDs:               s,
		Reserved:                      make(map[string]*UserSandbox),
		SandboxExpiry:                 DefaultSandboxExpiry,
		CleanupFrequency:              DefaultCleanupFrequency,
		InactivityExpirationThreshold: DefaultInactivityExpirationThreshold,
	}
}

func (s *Sandbox) AvailableCount() int {
	return len(s.AvailableBoxIDs)
}

// returns nil first time
// returns non-nil for second+ times
func (s *Sandbox) GetExistingSandbox(user_id string) (*UserSandbox, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.Reserved[user_id]
	return b, ok
}

func (s *Sandbox) UpdateUsed(user_id string) {
	b := s.Reserved[user_id]
	if b != nil {
		b.LastUsed = time.Now()
	}
}

// Reserve returns free sandbox id in a thread safe way
func (s *Sandbox) Reserve(uid string) (box *UserSandbox, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.AvailableBoxIDs) == 0 {
		return box, ErrSandboxBusy
	} else {
		var selected int
		for k := range s.AvailableBoxIDs {
			selected = k
			break
		}

		box := &UserSandbox{
			Id:        selected,
			ExpiresAt: time.Now().Add(s.SandboxExpiry),
			LastUsed:  time.Now(),
		}
		s.Reserved[uid] = box
		delete(s.AvailableBoxIDs, selected)
		return box, nil
	}
}

// Release puts back sandbox id in the pool to be used by other
func (s *Sandbox) Release(uid string) {
	s.mu.Lock()
	s.AvailableBoxIDs[s.Reserved[uid].Id] = struct{}{}
	delete(s.Reserved, uid)
	s.mu.Unlock()
}

func (s *Sandbox) initBox(boxId int) error {
	return exec.Command("isolate", "--init", fmt.Sprintf("-b %v", boxId)).Run()
}

func (s *Sandbox) deleteBox(boxId int) error {
	return exec.Command("isolate", "--cleanup", fmt.Sprintf("-b %v", boxId)).Run()
}

func (s *Sandbox) InitCleanup() {
	ticker := time.NewTicker(s.CleanupFrequency)
	go func() {
		for {
			<-ticker.C
			s.Cleanup()
		}
	}()
}

// Use this function when long running sandboxes start accumulating memory/cpu
func (s *Sandbox) Cleanup() {
	log.Println("sandbox cleanup started...")
	s.mu.Lock()
	toRelease := []string{}
	for userId, box := range s.Reserved {
		if box == nil {
			continue
		}

		expired := box.ExpiresAt.Before(time.Now())
		thresholdCrossed := box.LastUsed.Add(s.InactivityExpirationThreshold).Before(time.Now())
		if expired || thresholdCrossed {
			err := s.deleteBox(box.Id)
			if err != nil {
				log.Printf("failed to cleanup sandbox: %+v %v\n", *box, err)
				continue
			}

			err = s.initBox(box.Id)
			if err != nil {
				log.Println("error initializing box", box.Id)
				continue
			}

			// release deleted sandbox available
			toRelease = append(toRelease, userId)
		}
	}
	s.mu.Unlock()

	for i := 0; i < len(toRelease); i++ {
		s.Release(toRelease[i])
	}
	log.Println("total sandbox cleaned:", len(toRelease))
}
