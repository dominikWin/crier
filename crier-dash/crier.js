var message_container = document.getElementById("message_container");
var contents = "";

var websocket = new WebSocket("ws://localhost:8000/ws");

websocket.onopen = function () {
    websocket.send('init');
};

function message_to_table(msg) {
    return `<table class="ui celled very compact red table">
    <thead>
        <tr>
            <th>Time</th>
            <th>Host</th>
            <th>Message</th>
        </tr>
    </thead>
    <tbody>
        <tr>
        <td>${msg.id}</td>
        <td>${msg.host}</td>
        <td>${msg.message_head}</td>
        </tr>
    </tbody>
</table>`;
}

function add_message(msg) {
    contents = message_to_table(msg) + contents;
    message_container.innerHTML = contents;
}

websocket.onmessage = function (event) {
    add_message(JSON.parse(event.data))
};