# Go Sandbox Service

A containerized service that exposes a gRPC server streaming endpoint for running an arbitrary Go code in a sandbox.

The arbitrary code runs inside multiple sandboxes using [isolate](https://github.com/ioi/isolate).

<img src="https://github.com/nirdosh17/go-sandbox/assets/5920689/d71453d4-6843-42cf-a09e-23d668f6e72d" width="600" alt="sandbox arch" />


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

    You get real-time output from the executing code through the streaming endpoint, mirroring local execution.
    <video src="https://github.com/nirdosh17/go-sandbox/assets/5920689/cd67a4a3-ef32-43fa-bd33-969c34abc124" width="600" alt="sandbox api call demo" />

    **Request sample:**

    `session_id` can be used to bind a sandbox to a session(execution), e.g for authenticated users. If not provided, the code will run in random sandboxes.

    ```json
    {
      "code": "package main\n\nimport (\n\t\"fmt\"\n\t\"time\"\n)\n\nfunc main() {\n\tfor i := 0; i < 3; i++ {\n\t\ttime.Sleep(time.Second)\n\t\tfmt.Println(\"Hello\", i)\n\t}\n\n}\n",
      "session_id": "user_1" // optional
    }
    ```

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
    // error
    {
      "output": "main.go:10:8: undefined: time.Slseep",
      "exec_err": "",
      "is_error": false,
      "timestamp": "1712416529383"
    }
    ```
