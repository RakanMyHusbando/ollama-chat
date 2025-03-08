const chatHistory = document.querySelector(".chat-history");
const dropdownModels = document.querySelector(
    ".headline-elem.dropdown.dropdown-model",
);
const dropdownChats = document.querySelector(
    ".headline-elem.dropdown.dropdown-chats",
);
const messageInput = document.querySelector(".message-form input");
const messageButton = document.querySelector(".message-form button");

import { Ollama, Api, Chat, Message, makeId, ollamaUrl } from "./script.js";

const newChat = (api, userId) => {
    const c = new Chat(makeId(10), userId, new Date().toISOString(), [], api);
    localStorage.setItem("chat_id", c.content.id);
    return c;
};

const userId = parseInt(
    document.cookie.match(new RegExp(`(^| )user_id=([^;]+)`)).pop(),
);

const ollama = new Ollama(ollamaUrl());
const api = new Api();
/** @type {Chat} */
let chat;

if (!localStorage.getItem("chat_id")) chat = newChat(userId);
else
    await api.getChats(localStorage.getItem("chat_id"), true).then((chats) => {
        chat = chats.length > 0 ? chats[0] : newChat(userId);
    });

const runMessage = () => {
    if (messageInput.value.length > 0) {
        const msg = chat.addMessage(
            messageInput.value,
            "user",
            new Date().toISOString(),
        );
        msg.createHTML();
        console.log(msg.htmlElement);
        chatHistory.appendChild(msg.htmlElement);
        messageInput.value = "";
    }
};
messageButton.addEventListener("click", () => runMessage());
messageInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") runMessage();
});
