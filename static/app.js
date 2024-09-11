document.getElementById('connect').addEventListener('click', () => {
    const host = document.getElementById('host').value;
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    if (!host || !username || !password) {
        alert('Please fill in all fields.');
        return;
    }

    const connectInfo = {
        host: host,
        username: username,
        password: password
    }

    const msgType = {
        CONNECT: "CONNECT" ,
        COMMAND :  "COMMAND" ,
        DISCONNECT : "DISCONNECT"
    }

    const socket = new WebSocket('ws://localhost:8080/ws');

    socket.onopen = () => {
        const wsMsg = {
            type: msgType.CONNECT,
            data :connectInfo        
        }

        socket.send(JSON.stringify(wsMsg));
    };

    socket.onmessage = (event) => {
        document.getElementById('output').textContent += event.data + '\n';
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        alert('Connection error.');
    };

    document.getElementById('send-command').addEventListener('click', () => {
        const command = document.getElementById('command').value;

        const wsMsg = {
            type: msgType.COMMAND,
            data : command
        }

        if (command && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(wsMsg));
        }
    });
});
