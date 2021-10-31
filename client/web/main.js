let hub = 'earth';

function addMessage(msg) {
    const el = document.createElement('span');
    el.innerText = msg;
    document.getElementById("text-container").appendChild(el);
}

function setHub(value) {
    hub = value;
}

function leftpad(n) {
    if (n >= 10) return "" + n;
    return "0" + n;
}

function getTimestamp() {
    const d = new Date();
    return `${leftpad(d.getHours())}:${leftpad(d.getMinutes())}:${leftpad(d.getSeconds())}`
}

// setup validation for username input
const btn = document.getElementById('connect-button');
document.getElementById('username').oninput = (event) => {
    btn.disabled = event.target.value === '';
};

function start() {
    /*
    "Login"-modal
    */
    const name = document.getElementById('username').value;
    document.getElementById('modal').style.display = "none";
    
    /*
    Websocket Connection
    */
    const socket = new WebSocket(`ws://localhost:8080?name=${name}&hub=${hub}`);
    socket.addEventListener('message', function (event) {
        addMessage(`[rcvd ${getTimestamp()}] ` + event.data);
    });

    document.getElementById('username-text').innerHTML = name;
    document.getElementById('current-hub-text').innerHTML = hub;
        
        
    /*
    Input Keyhandler
    */
    document.getElementById('input').onkeydown = (event) => {
        if (event.keyCode === 13) {
            addMessage(`[sent ${getTimestamp()}] ${name}: ${event.target.value}`);
            socket.send(event.target.value);
            event.target.value = null;
        }
    }
}

/*
Create hub selection
*/
const hubSelection = document.getElementById('hub-selection');
function addHub(name) {
    const  cont = document.createElement('div');
    cont.classList.add('radio-container');

    const input = document.createElement('input');
    input.type = "radio";
    input.id = name;
    input.value = name;
    input.name = "hubs";
    input.onclick = () => hub = name;

    const label = document.createElement('label');
    label.htmlFor = name;
    label.innerHTML = name[0].toUpperCase() + name.replace(/-/g, ' ').slice(1); ;

    cont.appendChild(input);
    cont.appendChild(label);
    hubSelection.appendChild(cont);
}

const availableHubs = ['earth', 'mars', 'jupiter-station'];
availableHubs.forEach(addHub);
document.getElementById('earth').checked = true;