document.getElementById('connect').addEventListener('click', () => {
    const host = document.getElementById('host').value;
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    if (!host || !username || !password) {
        alert('Please fill in all fields.');
        return;
    }

    const socket = new WebSocket('ws://localhost:8080/ws');

    socket.onopen = () => {
        const sshInfo = JSON.stringify({ host, username, password });
        socket.send(sshInfo);
    };

    socket.onmessage = (event) => {
        document.getElementById('output').textContent = event.data;
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        alert('Connection error.');
    };
});
