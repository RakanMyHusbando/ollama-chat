const chatHistory = document.querySelector(".chat-history");
const dropdownModels = document.querySelector(
    ".headline-elem.dropdown.dropdown-model",
);
const dropdownChats = document.querySelector(
    ".headline-elem.dropdown.dropdown-chats",
);
const messageInput = document.querySelector(".message-form input");
const messageButton = document.querySelector(".message-form button");

import { Api, Chat, Message, makeId, ollamaUrl, userId } from "./script.js";

let chat = new Chat(ollamaUrl(), makeId(), userId());
const api = new Api();

const runMessage = async () => {
    if (messageInput.value.length > 0 && dropdownModels.value != "") {
        const msg = chat.addMessage("user", messageInput.value);
        chat.content.messages.length > 1 ? msg.post() : chat.post();
        chatHistory.appendChild(msg.createHTML());
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
    chats.forEach((chatId) => {
        const option = document.createElement("option");
        option.value = chatId;
        option.text = chatId;
        dropdownChats.appendChild(option);
    });
});

messageButton.addEventListener("click", () => runMessage());
messageInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") runMessage();
});

dropdownChats.addEventListener("change", async () => {
    chat = await api.getChats(dropdownChats.value, true);
    chatHistory.innerHTML = "";
    chat.content.messages.forEach((msg) =>
        chatHistory.appendChild(msg.createHTML()),
    );
});
