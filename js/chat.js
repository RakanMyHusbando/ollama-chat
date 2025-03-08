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
    const c = new Chat(makeId(), userId, new Date().toISOString(), [], api);
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

const runMessage = async () => {
    if (messageInput.value.length > 0) {
        const msg = chat.addMessage(
            messageInput.value,
            "user",
            new Date().toISOString(),
        );
        msg.createHTML();
        chatHistory.appendChild(msg.htmlElement);
        messageInput.value = "";
        console.log(chat.formJson().messages);
        const response = await ollama.chatStream(
            ollama.models[0],
            chat.formJson().messages,
        );
        const resMsg = new Message(
            localStorage.getItem("chat_id"),
            "",
            "assistant",
            new Date().toISOString(),
        );
        resMsg.createHTML();
        chatHistory.appendChild(resMsg.htmlElement);
        chat.addResStreamMessage(resMsg, response);
    }
};
messageButton.addEventListener("click", () => runMessage());
messageInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") runMessage();
});
