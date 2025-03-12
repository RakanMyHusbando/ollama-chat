const chatHistory = document.querySelector(".chat-history");
const dropdownModels = document.querySelector(
    ".headline-elem.dropdown.dropdown-model",
);
const dropdownChats = document.querySelector(
    ".headline-elem.dropdown.dropdown-chats",
);
const messageInput = document.querySelector(".message.form input");
const messageButton = document.querySelector(".message.form button");

import { Api, Chat, Message } from "./script.js";

let chat = new Chat();
const api = new Api();

const runMessage = async () => {
    if (messageInput.value.length > 0 && dropdownModels.value != "") {
        const msg = chat.addMessage("user", messageInput.value);
        if ([0, 1].includes(chat.content.messages.length % 10))
            chat.updateName();
        chatHistory.appendChild(msg.createHTML());
        chat.content.messages.length > 1 ? msg.post() : chat.post();
        messageInput.value = "";
        const assistantMsg = new Message(chat.content.id, "assistant");
        chatHistory.appendChild(assistantMsg.createHTML());
        chat.addOllamaMessage(assistantMsg, dropdownModels.value);
    }
};

chat.getModel().then(() =>
    chat.models.forEach((model) => {
        const option = document.createElement("option");
        option.value = model;
        option.text = model;
        dropdownModels.appendChild(option);
    }),
);

api.getChats().then((chats) => {
    chats.forEach((elem) => {
        const option = document.createElement("option");
        option.value = elem.content.id;
        option.text = elem.content.name;
        dropdownChats.appendChild(option);
    });
});

messageButton.addEventListener("click", () => runMessage());
messageInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") runMessage();
});

dropdownChats.addEventListener("change", async () => {
    chatHistory.innerHTML = "";
    if (dropdownChats.value == "") chat = new Chat();
    else
        api.getChats(dropdownChats.value, true).then((res) => {
            chat = res[0];
            chat.content.messages.forEach((msg) =>
                chatHistory.appendChild(msg.createHTML()),
            );
            chatHistory.scrollTop = chatHistory.scrollHeight;
        });
});
