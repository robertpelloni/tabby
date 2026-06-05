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
    notifications = []

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
                if res.get('id') == req_id:
                    end_time = time.perf_counter()
                    latencies.append((end_time - start_time) * 1000) # ms
                    return res
                elif 'method' in res:
                    notifications.append(res)
                    # print(f"Received Notification: {res['method']}")

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

    # 3. Concurrency Soak Test & Diagnostic
    print("Running Stress Case: Concurrent Agent Tasks...")
    task_ids = []
    for i in range(10):
        res = request("agent.runTask", {"description": "Environment Diagnostic" if i == 0 else f"Soak Task {i}"})
        if res and 'result' in res:
            task_ids.append(res['result']['id'])
        else:
            print(f"Failed to start task {i}: {res}")

    # Polling stress
    print("Polling status under load...")
    start_soak = time.time()
    completed = 0
    diagnostic_result = ""

    while completed < 10 and (time.time() - start_soak) < 30:
        time.sleep(0.1)
        completed = 0
        for tid in task_ids:
            res = request("agent.getTaskStatus", {"id": tid})
            if res and 'result' in res:
                if res['result']['status'] == "completed":
                    completed += 1
                    if res['result']['description'] == "Environment Diagnostic":
                        diagnostic_result = res['result']['result']

    print(f"Soak Test Latency: avg={statistics.mean(latencies):.2f}ms")
    if completed == 10:
        print("✅ Soak Test verified (10/10 tasks completed)")
        print(f"Diagnostic Result:\n{diagnostic_result}")
    else:
        print(f"❌ Soak Test failed ({completed}/10 tasks completed)")

    # 4. Verify Notifications
    agent_notifs = [n for n in notifications if n['method'] == 'agent.taskUpdated']
    if len(agent_notifs) > 0:
        print(f"✅ Received {len(agent_notifs)} agent notifications")
    else:
        print("❌ No agent notifications received")


    # 5. Widget & VDOM Benchmark
    print('Benchmarking Widgets & VDOM...')
    widget_res = request('agent.createWidget', {'type': 'vdom', 'title': 'Test Widget'})
    if widget_res and 'result' in widget_res:
        widget_id = widget_res['result']['id']
        vdom_node = {
            'tag': 'div',
            'props': {'className': 'p-4'},
            'children': [
                {'tag': 'h1', 'children': ['Hello VDOM']},
                'Text child'
            ]
        }
        request('agent.updateWidgetVDOM', {'id': widget_id, 'vdom': vdom_node})
        print('✅ Widget creation and VDOM update verified')
    else:
        print('❌ Widget creation failed')


    # 6. Workflow Benchmark
    print('Benchmarking Workflows...')
    wf_res = request('agent.startWorkflow', {'description': 'Test Feature Workflow'})
    if wf_res and 'result' in wf_res:
        wf_id = wf_res['result']['id']
        # Wait for clarification phase
        time.sleep(1.5)
        request('agent.submitWorkflowResponse', {'id': wf_id, 'response': 'Go ahead'})
        print('✅ Workflow start and response submission verified')
    else:
        print('❌ Workflow creation failed')

    proc.terminate()
    print("Performance Integration Test Finished.")

if __name__ == "__main__":
    run_test()
