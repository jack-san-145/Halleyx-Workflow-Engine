// Common functions
async function fetchJSON(url) {
    const res = await fetch(url);
    if (!res.ok) throw new Error(await res.text());
    return res.json();
}

// Load workflows on index page
if (window.location.pathname === '/frontend/' || window.location.pathname === "/index_page") {
    loadWorkflows();
}

async function loadWorkflows() {
    try {
        const workflows = await fetchJSON('/api/workflows');
        const tbody = document.querySelector('#workflows tbody');
        tbody.innerHTML = '';
        workflows.forEach(wf => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
                <td>${wf.name}</td>
                <td>${wf.version}</td>
                <td>${wf.is_active ? 'Yes' : 'No'}</td>
                <td>
                    <button onclick="startExecution('${wf.id}')">Execute</button>
                    <button onclick="viewExecution('${wf.id}')">View Executions</button>
                </td>
            `;
            tbody.appendChild(tr);
        });
    } catch (err) {
        alert('Failed to load workflows: ' + err);
    }
}

async function startExecution(workflowID) {
    const data = prompt('Enter execution data as JSON (e.g., {"code":"int main(){return 0;}"})');
    if (!data) return;
    try {
        const payload = JSON.parse(data);
        const res = await fetch(`/api/workflows/${workflowID}/execute`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ data: payload })
        });
        if (!res.ok) throw new Error(await res.text());
        const exec = await res.json();
        alert('Execution started: ' + exec.id);
        window.location.href = `/execution_page?id=${exec.id}`;
    } catch (err) {
        alert('Error: ' + err);
    }
}

// Execution detail page with WebSocket
if (window.location.pathname === '/execution_page') {
    const urlParams = new URLSearchParams(window.location.search);
    const execId = urlParams.get('id');
    if (execId) {
        initExecutionPage(execId);
    }
}

function initExecutionPage(execId) {
    let socket = null;
    let executionData = null;
    let steps = [];

    // Fetch initial execution data
    fetchJSON(`/api/executions/${execId}`).then(data => {
        executionData = data;
        document.getElementById('exec-id').innerText = data.id;
        document.getElementById('exec-status').innerText = data.status;
        document.getElementById('workflow-name').innerText = data.workflow_id; // could fetch workflow name separately
        // Fetch steps? For now, we'll get from workflow, but we need steps list.
        // For simplicity, we'll assume the UI will update via WebSocket.
    }).catch(console.error);

    // Connect WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    socket = new WebSocket(`${protocol}//${window.location.host}/api/ws`);

    socket.onopen = () => {
        console.log('WebSocket connected');
    };

    socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        handleWebSocketMessage(msg);
    };

    socket.onerror = (err) => console.error('WebSocket error', err);
    socket.onclose = () => console.log('WebSocket closed');

    function handleWebSocketMessage(msg) {
        console.log('Received:', msg);
        if (msg.payload.execution_id !== execId) return;

        switch (msg.type) {
            case 'STEP_STARTED':
                updateStepStatus(msg.payload.step_id, 'in_progress');
                break;
            case 'STEP_COMPLETED':
                updateStepStatus(msg.payload.step_id, 'completed');
                break;
            case 'STEP_FAILED':
                updateStepStatus(msg.payload.step_id, 'failed');
                break;
            case 'EXECUTION_COMPLETED':
                document.getElementById('exec-status').innerText = 'completed';
                break;
        }
    }

    function updateStepStatus(stepId, status) {
        const stepDiv = document.getElementById(`step-${stepId}`);
        if (stepDiv) {
            stepDiv.className = `step-card step-${status}`;
            const statusSpan = stepDiv.querySelector('.step-status');
            if (statusSpan) statusSpan.innerText = status;
        }
    }

    // For demo, we'll create placeholder steps after a short delay
    setTimeout(() => {
        // Create step cards based on initial execution data? Actually we need steps from workflow.
        // We'll just create a few dummy steps to demonstrate.
        const stepsContainer = document.getElementById('steps');
        stepsContainer.innerHTML = '';
        const stepIds = ['step1', 'step2', 'step3'];
        stepIds.forEach((id, idx) => {
            const div = document.createElement('div');
            div.id = `step-${id}`;
            div.className = 'step-card step-pending';
            div.innerHTML = `<h3>Step ${idx+1}</h3><p>Status: <span class="step-status">pending</span></p>`;
            stepsContainer.appendChild(div);
        });
    }, 1000);
}