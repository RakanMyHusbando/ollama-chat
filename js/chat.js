const chatHistory = document.querySelector(".chat-history");
const dropdownModels = document.querySelector(
    ".headline-elem.dropdown.dropdown-model",
);
const dropdownChats = document.querySelector(
    ".headline-elem.dropdown.dropdown-chats",
);
const messageInput = document.querySelector(".message-form input");
const messageButton = document.querySelector(".message-form button");

const ollamaUrl = "http://localhost:11434";

let chatId = 0;
const context = [];

const appendMessage = (role, message) => {
    const messageElem = document.createElement("div");
    const textElem = document.createElement("div");
    messageElem.classList.add("message");
    messageElem.classList.add(role);
    textElem.classList.add("text");
    textElem.innerText = message;
    messageElem.appendChild(textElem);
    chatHistory.appendChild(messageElem);
    return messageElem;
};

const updateMessage = (elem, message) => {
    elem.children[0].innerText += message;
};

const appendModel = (model) => {
    const optionElem = document.createElement("option");
    optionElem.innerText = model;
    optionElem.value = model;
    dropdownModels.appendChild(optionElem);
    return optionElem;
};

const chat = async (question, model) => {
    let role = "user";
    if (question === "") {
        throw new Error("Question is empty");
    }
    context.push({ role, content: question });
    const response = await fetch(ollamaUrl + "/api/chat", {
        method: "POST",
        body: JSON.stringify({
            model: model,
            messages: context,
        }),
    });
    if (!response.ok) {
        throw new Error(
            `HTTP error! Status: ${response.status} ${response.statusText}`,
        );
    }
    if (!response.body) {
        throw new Error("Readable stream not found in the response.");
    }
    let text = "";
    const reader = response.body.getReader();
    const assistMsgElem = appendMessage("assistant", text);
    while (true) {
        const { done, value } = await reader.read();
        if (done) {
            context.push({ role: "assistant", content: text });
            return;
        }
        const decodeStr = new TextDecoder().decode(value);
        const res = JSON.parse(decodeStr);
        text = text + res.message.content;
        updateMessage(assistMsgElem, res.message.content);
    }
};

const tags = async () => {
    const response = await fetch(ollamaUrl + "/api/tags");
    if (!response.ok) {
        throw new Error(
            `HTTP error! Status: ${response.status} ${response.statusText}`,
        );
    }
    const data = await response.json();
    const models = [];
    for (const model of data.models) {
        models.push(model.name.split(":")[0]);
    }
    return models;
};

const runMessage = async () => {
    const msg = messageInput.value;
    messageInput.value = "";
    appendMessage("user", msg);
    console.log(dropdownModels.value);
    await chat(msg, dropdownModels.value).catch((err) => console.error(err));
};

tags().then((res) => res.forEach((elem) => appendModel(elem)));

messageButton.addEventListener("click", () => runMessage());
messageInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") runMessage();
});
