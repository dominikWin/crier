var message_container = document.getElementById("message_container");
var old_contents = "";
var active_host = "";
var active_messages = [];
var colors = ["red", "blue", "orange", "green", "violet", "black", "yellow", "teal", "pink", "olive"];
var color_counter = 0;
var host_colors = new Map();
var last_message_id = "0-0"

function get_host_color(host) {
    if(host_colors.has(host)) {
        return host_colors.get(host);
    }

    var color = colors[color_counter];
    color_counter++;
    if(color_counter >= colors.length) {
        color_counter = 0;
    }

    host_colors.set(host, color);
    return color;
}

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
            old_contents = messages_to_table(active_messages, get_host_color(active_host)) + old_contents;
        }
        active_messages = [];
    }
    active_host = msg.host;
    active_messages = active_messages.concat([msg]);
    last_message_id = msg.id;
    var contents = messages_to_table(active_messages, get_host_color(active_host)) + old_contents;
    message_container.innerHTML = contents;
}

function error_dim(dim) {
    if(dim) {
        $('body > .page.dimmer').dimmer('show');
        $('body > .page.dimmer').dimmer.settings.closable = false;
    } else {
        $('body > .page.dimmer').dimmer('hide');
    }
}

function attempt_ws_start() {
    var websocket = new WebSocket(`ws://${window.location.host}/ws`);

    websocket.onopen = function () {
        websocket.send(last_message_id);
        error_dim(false);
    };
    
    websocket.onmessage = function (event) {
        add_message(JSON.parse(event.data));
    };

    websocket.onerror = function(event) {
        websocket.close()
    }

    websocket.onclose = function(event) {
        error_dim(true);
        setTimeout(attempt_ws_start, 3000);
    }
}

$(document).ready(attempt_ws_start);

