const node = (new URLSearchParams(document.location.search)).get("name")

const nameH = document.getElementById("name-h");
nameH.innerText = `Node name: ${node}`

const addrInput = document.getElementById("addr-input");
const directButton = document.getElementById("direct-button")

const natInput = document.getElementById("nat-input");
const natButton = document.getElementById("nat-button")

const destInput = document.getElementById("dest-input");
const textInput = document.getElementById("text-input")
const sendButton = document.getElementById("send-button")

const refreshP = document.getElementById("refresh-p");
const refreshButton = document.getElementById("refresh-button");

const localP = document.getElementById("local-p")

const neighborList = document.getElementById("neighbor-list")
const addrList = document.getElementById("addr-list");
const routingList = document.getElementById("routing-list");
const chatList = document.getElementById("chat-list")

function appendToNodeList(text, list) {
  let li = document.createElement("li");

  let textNode = document.createTextNode(text);

  li.appendChild(textNode);
  list.appendChild(li);
}

const api = `/api/nodes/${node}`

function fetchNodeData() {
  fetch(api).then(resp => resp.json()).then(data => {
    localP.innerText = `Local addr: ${data.local}`

    neighborList.innerHTML = ""
    for (const key in data.neigh) {
      appendToNodeList(`name: ${key}, addr: ${data.neigh[key]}`, neighborList)
    }

    addrList.innerHTML = ""
    for (const addr of data.addr) {
      appendToNodeList(`${addr}`, addrList)
    }

    routingList.innerHTML = ""
    for (const key in data.routing) {
      appendToNodeList(`for ${key}, go to ${data.routing[key]}`, routingList)
    }

    chatList.innerHTML = ""
    for (const msg of data.chat) {
      appendToNodeList(`${msg.text} | from ${msg.src} | at ${msg.time}`, chatList)
    }
  });
}

let tick = 0;
const period = 5;
const onTick = () => {
  tick++;
  if (tick == period) {
    tick = 0;
    fetchNodeData();
  }
  updateRefreshP()
}

let timer = setInterval(onTick, 1000);

fetchNodeData();
updateRefreshP();

function updateRefreshP() {
  refreshP.innerText = `Automatic refresh in ${5 - tick} sec`
}

refreshButton.onclick = () => {
  fetchNodeData();
  tick = 0;
  updateRefreshP();
}

directButton.onclick = () => {
  let addr = addrInput.value
  postData(api, { op: "direct", data: { addr: addr }});
  fetchNodeList();
}

natButton.onclick = () => {
  let dest = natInput.value
  postData(api, { op: "nat", data: { dest: dest }});
  fetchNodeList();
}

sendButton.onclick = () => {
  let dest = destInput.value
  let text = textInput.value
  postData(api, { op: "chat", data: { dest: dest, text: text }});
  fetchNodeList();
}
