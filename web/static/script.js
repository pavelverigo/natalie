const nameInput = document.getElementById("name-input");
const portInput = document.getElementById("port-input");
const addButton = document.getElementById("add-button");

const refreshP = document.getElementById("refresh-p");
const refreshButton = document.getElementById("refresh-button");

const nodeList = document.getElementById("node-list");

function onlyLettersAndNumbers(str) {
  return /^[A-Za-z0-9]+$/.test(str);
}

function appendToNodeList(name) {
  let li = document.createElement("li");

  let a = document.createElement("a");
  let nameText = document.createTextNode(name);
  a.href = `/nodes/?name=${name}`;
  a.appendChild(nameText);

  // let stop = document.createElement("button");
  // let stopText = document.createTextNode("Stop");
  // stop.appendChild(stopText)

  li.appendChild(a);
  // li.appendChild(stop);
  nodeList.appendChild(li);
}

function fetchNodeList() {
  fetch("/api/nodes/").then(resp => resp.json()).then(data => {
    nodeList.innerHTML = "";
    for (const name of data) {
      appendToNodeList(name)
    }
  });
}

addButton.onclick = () => {
  let name = nameInput.value;
  let port = parseInt(portInput.value);
  if (onlyLettersAndNumbers(name)) {
    postData('/api/nodes/', { name: name, port: port });
    fetchNodeList();
  } else {
    console.log(`illegal name ${name}, use only letters and digits`)
  }
}

let tick = 0;
const period = 5;
const onTick = () => {
  tick++;
  if (tick == period) {
    tick = 0;
    fetchNodeList();
  }
  updateRefreshP()
}

let timer = setInterval(onTick, 1000);

fetchNodeList();
updateRefreshP();

function updateRefreshP() {
  refreshP.innerText = `Automatic refresh in ${5 - tick} sec`
}

refreshButton.onclick = () => {
  fetchNodeList();
  tick = 0;
  updateRefreshP();
}