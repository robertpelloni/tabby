import subprocess
import json
import time
import os
import threading

def run_test():
    print("Starting Comprehensive Integration & Edge Case Test...")
    backend_path = './build/tabby-backend'

    # Start backend
    proc = subprocess.Popen([backend_path], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

    def request(method, params=None, req_id=None):
        if req_id is None:
            req_id = int(time.time() * 1000)
        req = {"jsonrpc": "2.0", "id": req_id, "method": method, "params": params or {}}
        # print(f"-> Request: {method}")
        proc.stdin.write(json.dumps(req) + '\n')
        proc.stdin.flush()

        while True:
            line = proc.stdout.readline()
            if not line:
                return None
            if line.startswith('{'):
                return json.loads(line)

    # 1. Basic Connectivity
    res = request("ping")
    if res and 'result' in res:
        print(f"✅ Ping success: {res['result']['status']}")
    else:
        print("❌ Ping failed")

    # 2. Edge Case: Malformed JSON
    print("Testing Edge Case: Malformed JSON...")
    proc.stdin.write('{"jsonrpc": "2.0", "method": "ping", "params": { "broken": }\n')
    proc.stdin.flush()
    line = proc.stdout.readline()
    res = json.loads(line)
    if res.get('error', {}).get('code') == -32700:
        print("✅ Correctly handled Parse Error (-32700)")
    else:
        print(f"❌ Failed to handle Parse Error: {res}")

    # 3. Edge Case: Method Not Found
    print("Testing Edge Case: Method Not Found...")
    res = request("nonExistentMethod")
    if res.get('error', {}).get('code') == -32601:
        print("✅ Correctly handled Method Not Found (-32601)")
    else:
        print(f"❌ Failed to handle Method Not Found: {res}")

    # 4. Persistence Test (Sync Push -> Pull)
    print("Testing Sync Persistence...")
    test_data = {
        "workflows": [{"id": "test-wf", "name": "Test Workflow", "command": "echo hello", "tags": ["test"]}],
        "profiles": [],
        "envVars": []
    }
    request("sync.push", {"data": test_data})
    res = request("sync.pull")
    if res['result']['data']['workflows'][0]['id'] == "test-wf":
        print("✅ Sync Persistence verified")
    else:
        print("❌ Sync Persistence failed")

    # 5. Stress Test: Concurrent Agent Tasks
    print("Testing Stress Case: Concurrent Agent Tasks...")
    task_ids = []
    for i in range(5):
        res = request("agent.runTask", {"description": f"Concurrent Task {i}"})
        task_ids.append(res['result']['id'])

    print(f"Started 5 tasks: {task_ids}")

    # Wait and check all
    completed = 0
    for _ in range(20):
        time.sleep(0.5)
        completed = 0
        for tid in task_ids:
            res = request("agent.getTaskStatus", {"id": tid})
            if res['result']['status'] == "completed":
                completed += 1
        if completed == 5:
            break

    if completed == 5:
        print("✅ Concurrent Task Execution verified")
    else:
        print(f"❌ Concurrent Task Execution failed (only {completed}/5 completed)")

    # 6. Edge Case: Invalid Params for known method
    print("Testing Edge Case: Invalid Params...")
    # agent.getTaskStatus expects {"id": "..."}
    res = request("agent.getTaskStatus", {"wrong_key": "123"})
    if res.get('error'):
        print(f"✅ Correctly returned error for invalid params: {res['error']['message']}")
    else:
        print("❌ Failed to catch invalid params")

    proc.terminate()
    print("Integration Test Finished.")

if __name__ == "__main__":
    run_test()
