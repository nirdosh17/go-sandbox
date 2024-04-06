# Go Sandbox Service
Containerized service which expose gRPC endpoint to run arbitrary Go code.

The code is run inside sandbox using [isolate](https://github.com/ioi/isolate) package.

## Running locally
1. **Build image**
    ```bash
    make build
    ```
2. **Run gRPC service (server streaming)**
    ```bash
    # starts service in localhost:8080
    make run
    ```
3. **Make RPC call to execute arbitrary code**

    You get live output without need to wait for the code to finish executing.
    <video src="https://github.com/nirdosh17/go-sandbox/assets/5920689/cd67a4a3-ef32-43fa-bd33-969c34abc124" width="600" alt="sandbox api call demo" />

    **Response Stream:**

    ```json
    // success
    {
      "output": "Hello",            // stdout/stderr from executed Go code
      "exec_err": "",               // server error
      "is_error": false,            // true for server error
      "timestamp": "1712415917223"  // stdout/err timestamp
    }
    ```
    ```json
    // error from go code
    {
      "output": "/app/main.go:10:8: undefined: time.Slseep",
      "exec_err": "",
      "is_error": false,
      "timestamp": "1712416529383"
    }
    ```
