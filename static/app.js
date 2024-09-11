// 페이지 로드 시 저장된 접속 정보 불러오기
window.addEventListener('load', () => {
    const savedHost = localStorage.getItem('host');
    const savedUsername = localStorage.getItem('username');
    const savedPassword = localStorage.getItem('password');

    if (savedHost) document.getElementById('host').value = savedHost;
    if (savedUsername) document.getElementById('username').value = savedUsername;
    if (savedPassword) document.getElementById('password').value = savedPassword;
});

document.getElementById('connect').addEventListener('click', () => {
    const host = document.getElementById('host').value;
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    if (!host || !username || !password) {
        alert('Please fill in all fields.');
        return;
    }

    // 접속 정보를 Local Storage에 저장
    localStorage.setItem('host', host);
    localStorage.setItem('username', username);
    localStorage.setItem('password', password);

    const connectInfo = {
        host: host,
        username: username,
        password: password
    };

    const msgType = {
        CONNECT: "CONNECT",
        COMMAND: "COMMAND",
        DISCONNECT: "DISCONNECT"
    };

    const socket = new WebSocket('ws://localhost:8080/ws');

    socket.onopen = () => {
        const wsMsg = {
            type: msgType.CONNECT,
            data: connectInfo
        };

        socket.send(JSON.stringify(wsMsg));
    };

    socket.onmessage = (event) => {
        const output = document.getElementById('output');
        const messageElement = document.createElement('div');
        messageElement.textContent = event.data;
        output.appendChild(messageElement);

        messageElement.scrollIntoView({ behavior: "smooth", block: "end" });

        if (event.data.includes("Connected to SSH server")) {
            document.getElementById('connection-container').classList.add('hidden');
            document.getElementById('terminal-container').classList.remove('hidden');
        }
    };

    socket.onerror = (error) => {
        console.error('WebSocket error:', error);
        alert('Connection error.');
    };

    document.getElementById('command').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            const command = e.target.value.trim();
            if (command.toLowerCase() === 'clear') {
                document.getElementById('output').innerHTML = '';
                e.target.value = '';
                return;
            }

            if (command) {
                const wsMsg = {
                    type: msgType.COMMAND,
                    data: command
                };
                socket.send(JSON.stringify(wsMsg));
                e.target.value = '';
            }
        }
    });
});
