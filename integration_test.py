import subprocess
import json
import time
import os

def run_test():
    print("Starting Integration Test...")
    backend_path = './build/tabby-backend'

    # Start backend
    proc = subprocess.Popen([backend_path], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

    def request(method, params=None):
        req_id = int(time.time() * 1000)
        req = {"jsonrpc": "2.0", "id": req_id, "method": method, "params": params or {}}
        print(f"-> Request: {method}")
        proc.stdin.write(json.dumps(req) + '\n')
        proc.stdin.flush()

        while True:
            line = proc.stdout.readline()
            if not line:
                return None
            if line.startswith('{'):
                return json.loads(line)

    # 1. Ping
    res = request("ping")
    print(f"Ping success: {res['result']}")

    # 2. SFTP Upload (Simulation)
    res = request("sftp.upload", {
        "sessionId": "fake-sess",
        "remotePath": "/tmp/test",
        "localPath": "/etc/hosts",
        "transferId": "test-id-123"
    })
    print(f"SFTP Upload Response (expected error): {res.get('error', {}).get('message')}")

    # 3. AI Chat (Simulation)
    res = request("ai.chat", {"message": "hello"})
    print(f"AI Chat Response: {res['result']['response']}")

    # 4. Sync Pull (Simulation)
    res = request("sync.pull")
    print(f"Sync Pull Response: {res['result']['timestamp']}")

    # 5. Agent Task Execution
    print("Testing Agent Protocol...")
    res = request("agent.runTask", {"description": "Verify Integration"})
    task_id = res['result']['id']
    print(f"Agent Task Started: {task_id}")

    res = request("agent.listTasks")
    print(f"Agent List Tasks: {len(res['result'])} tasks found")

    # Wait for completion
    print("Waiting for task completion...")
    for _ in range(10):
        time.sleep(0.5)
        res = request("agent.getTaskStatus", {"id": task_id})
        status = res['result']['status']
        progress = res['result']['progress']
        print(f"Task {task_id} status: {status} ({progress*100:.0f}%)")
        if status == "completed":
            break

    proc.terminate()
    print("Integration Test Finished.")

if __name__ == "__main__":
    run_test()
