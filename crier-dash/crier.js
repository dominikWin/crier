var message_container = document.getElementById("message_container");
var old_contents = "";
var active_host = "";
var active_messages = [];

var websocket = new WebSocket("ws://localhost:8000/ws");

websocket.onopen = function () {
    websocket.send('init');
};

function messages_to_table(msg, color) {
    var rows = "";
    for(var i = 0; i < msg.length; i++) {
        rows = `<tr>
        <td>${msg[i].id}</td>
        <td>${msg[i].host}</td>
        <td>${msg[i].message_head}</td>
        </tr>` + rows;
    }
    return `<table class="ui celled very compact ${color} table">
    <thead>
        <tr>
            <th width="10%">ID</th>
            <th width="12%">Host</th>
            <th width="78%">Message</th>
        </tr>
    </thead>
    <tbody>
        ${rows}
    </tbody>
</table>`;
}

function add_message(msg) {
    if(msg.host != active_host) {
        if(active_messages.length > 0) {
            old_contents = messages_to_table(active_messages) + old_contents;
        }
        active_messages = [];
    }
    active_host = msg.host;
    active_messages = active_messages.concat([msg]);
    var contents = messages_to_table(active_messages) + old_contents;
    message_container.innerHTML = contents;
}

websocket.onmessage = function (event) {
    add_message(JSON.parse(event.data))
};