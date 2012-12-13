var header = "[wstail]"
var ws;
function init() {
  console.log("init");
  if (ws != null) {
    ws.close();
    ws = null;
  }
  path = "/tail";
  console.log("path:" + path);
  var div = document.getElementById("tail");
  div.innerText = div.innerText + header + "path:" + path + "\n";
  ws = new WebSocket("ws://localhost:23456" + path);
  ws.onopen = function () {
    div.innerText = div.innerText + header + "opened\n";
  };
  ws.onmessage = function (e) {
    console.log(e.data);
    var data = JSON.parse(e.data);
    if (data.key == "filename") {
      document.getElementById("info").innerText = data.value;
    } else if (data.key == "msg") {
      div.innerText = div.innerText + data.value;
    }
  };
  ws.onclose = function (e) {
    div.innerText = div.innerText + header + "closed\n";
  };
  console.log("init");
  div.innerText = div.innerText + header + "init\n";
};
