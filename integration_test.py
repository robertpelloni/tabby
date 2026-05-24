import subprocess
import json
import time
import os
import statistics

def run_test():
    print("Starting Comprehensive Integration & Performance Benchmark...")
    backend_path = './build/tabby-backend'

    # Start backend
    proc = subprocess.Popen([backend_path], stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

    latencies = []

    def request(method, params=None, req_id=None):
        if req_id is None:
            req_id = int(time.time() * 1000)
        req = {"jsonrpc": "2.0", "id": req_id, "method": method, "params": params or {}}

        start_time = time.perf_counter()
        proc.stdin.write(json.dumps(req) + '\n')
        proc.stdin.flush()

        while True:
            line = proc.stdout.readline()
            if not line:
                return None
            if line.startswith('{'):
                res = json.loads(line)
                end_time = time.perf_counter()
                latencies.append((end_time - start_time) * 1000) # ms
                return res

    # 1. Connectivity & Baseline Latency
    print("Benchmarking Baseline (Ping)...")
    for _ in range(50):
        request("ping")

    print(f"Ping Latency: avg={statistics.mean(latencies):.2f}ms, min={min(latencies):.2f}ms, max={max(latencies):.2f}ms")
    latencies.clear()

    # 2. Sync Persistence Benchmark
    print("Benchmarking Sync Persistence (Push/Pull)...")
    test_data = {
        "workflows": [{"id": f"wf-{i}", "name": f"Workflow {i}", "command": "echo hello"} for i in range(100)],
        "profiles": [],
        "envVars": []
    }
    for _ in range(10):
        request("sync.push", {"data": test_data})
        request("sync.pull")

    print(f"Sync Latency (100 items): avg={statistics.mean(latencies):.2f}ms")
    latencies.clear()

    # 3. Concurrency Soak Test
    print("Running Stress Case: Concurrent Agent Tasks...")
    task_ids = []
    for i in range(10):
        res = request("agent.runTask", {"description": f"Soak Task {i}"})
        task_ids.append(res['result']['id'])

    # Polling stress
    print("Polling status under load...")
    start_soak = time.time()
    completed = 0
    while completed < 10 and (time.time() - start_soak) < 30:
        time.sleep(0.1)
        completed = 0
        for tid in task_ids:
            res = request("agent.getTaskStatus", {"id": tid})
            if res['result']['status'] == "completed":
                completed += 1

    print(f"Soak Test Latency: avg={statistics.mean(latencies):.2f}ms")
    if completed == 10:
        print("✅ Soak Test verified (10/10 tasks completed)")
    else:
        print(f"❌ Soak Test failed ({completed}/10 tasks completed)")

    proc.terminate()
    print("Performance Integration Test Finished.")

if __name__ == "__main__":
    run_test()
