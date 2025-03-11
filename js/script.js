const ollamaUrl = () => "http://217.160.124.151:11434";

const makeId = () => Math.floor(Math.random() * Date.now()).toString(36);

class Ollama {
    /** @type {string} */
    baseUrl;

    /** @type {string[]} */
    models = [];

    /** @param {string} baseUrl */
    constructor(baseUrl) {
        this.baseUrl = baseUrl;
        this.#getModel(this.#url("tags")).then(() => console.log(this.models));
    }

    /** @param {... string} apiPath */
    #url = (...apiPath) => `${this.baseUrl}/api/${apiPath.join("/")}`;

    #getModel = async () => {
        const res = await fetch(this.#url("tags"));
        const data = await res.json();
        data.models.forEach((model) => this.models.push(model.name));
    };

    #httpError = (status, statusText) =>
        new Error(`HTTP Error! Status: ${status}: ${statusText}`);

    /**
     * @param {string} model
     * @param {{role: string, content: string}[]} messages
     * @returns {Promise<Response>}
     */
    chatStream = async (model, messages) => {
        try {
            console.log({ model, messages });
            const res = await fetch(this.#url("chat"), {
                method: "POST",
                body: JSON.stringify({ model, messages }),
            });
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            if (!res.body) throw new Error("Readable stream not found.");
            return res;
        } catch (error) {
            console.error(error);
        }
    };

    generate = async (model, prompt) => {
        try {
            const res = await fetch(this.#url("generate"), {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ model, prompt, stream: false }),
            });
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            return res.json();
        } catch (error) {
            console.error(error);
        }
    };

    /**
     * Pulls a stream of data from the server.
     * @param {string} model - The model to use for the stream.
     * @returns {Response} - The reader for the stream.
     */
    pullStream = async (model) => {
        try {
            const res = await fetch(this.#url(["pull"]), {
                method: "POST",
                body: JSON.stringify({ model }),
            });
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            if (!res.body) throw new Error("Readable stream not found.");
            return res;
        } catch (error) {
            console.error(error);
        }
    };
}

class Api {
    #httpError = (status, statusText) =>
        new Error(`HTTP Error! Status: ${status}: ${statusText}`);

    /** @param {Chat} chat */
    postChat = async (chat) => {
        try {
            const res = await fetch("/api/chat", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(chat.formJson()),
            });
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            return await res.json();
        } catch (error) {
            console.error(error);
        }
    };

    /**
     * @param {string|null} chatId
     * @param {true} [loadMessages]
     * @returns {Promise<Chat[]>}
     */
    getChats = async (chatId, loadMessages) => {
        try {
            const result = [];
            let query = chatId ? "?id=" + chatId : "";
            query += loadMessages ? (query != "" ? "&" : "?") + "msg=true" : "";
            const res = await fetch(`/api/chat${query}`);
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            const data = await res.json();
            data.forEach((chat) =>
                result.push(
                    new Chat(
                        chat.id,
                        chat.user_id,
                        chat.created_at,
                        chat.messages.map(
                            (msg) =>
                                new Message(
                                    msg.chat_id,
                                    msg.content,
                                    msg.role,
                                    msg.created_at,
                                ),
                        ),
                        chat.name,
                    ),
                ),
            );
            return result;
        } catch (error) {
            console.error(error);
        }
    };

    /** @param {Chat} chat */
    updateChatName = async (chat) => {
        try {
            const res = await fetch("/api/chat?change=name", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(chat.formJson()),
            });
            if (!res.ok) throw this.#httpError(res.status, res.statusText);
            return await res.json();
        } catch (error) {
            console.log(error);
        }
    };
}

class Chat {
    /** @type {{name: string, userId: string, createdAt: string, messages: Message[] }} */
    content;
    /** @type {Api} */
    api;

    /**
     * @param {string} id
     * @param {string} userId
     * @param {string} createdAt
     * @param {Message[]} messages
     * @param {string} [name="new chat"]
     */
    constructor(id, userId, createdAt, messages, name = "new chat") {
        this.content = { id, name, userId, createdAt, messages };
        this.api = new Api();
    }

    updateName = async () => await this.api.updateChatName(this);

    /**
     * @param {string} content - message content
     * @param {string} role - message role (user/assist)
     * @param {string} createdAt - message creation time
     * @returns {Message} - The message object.
     */
    addMessage = (content, role, createdAt) => {
        const msg = new Message(this.content.id, content, role, createdAt);
        this.content.messages.push(msg);
        return msg;
    };

    /**
     * @param {Message} message
     * @param {Response} response
     */
    addResStreamMessage = async (message, response) => {
        const badRes = ["", "<think>", "</think>"];
        const reader = response.body.getReader();
        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done) {
                    this.content.messages.push(message);
                    return;
                }
                new TextDecoder()
                    .decode(value)
                    .split("\n")
                    .forEach((e) => {
                        if (!badRes.includes(res.message.content)) {
                            const res = JSON.parse(e);
                            message.addText(res.message.content);
                        }
                    });
            }
        } catch (error) {
            console.error(error);
        }
    };

    /** @returns {Object} The JSON representation of the chat object.*/
    formJson = () => {
        const { id, name, user_id, created_at, messages } = this.content;
        return {
            id,
            name,
            user_id,
            created_at,
            messages: messages.map((msg) => msg.formJson()),
        };
    };

    /** @returns {string} */
    newNamePrompt = () => {
        const task =
            "Create a name for a chat.With maximum length of 3 words.Only respond with the name.";
        let chat = "";
        this.content.messages.forEach((msg) => {
            chat += `\t${msg.content.role}: ${msg.content.content}\n`;
        });
        return `<task>\n\t${task}\n</task>\n<chat>\n${chat}</chat>`;
    };
}

class Message {
    /** @type {{ chatId: string, content: string, role: string, createdAt: string }} */
    content;
    /** @type {HTMLDivElement} */
    htmlElement;
    /** @type {Api} */
    api;

    /**
     * @param {string} chatId
     * @param {string} content
     * @param {string} role
     * @param {string} createdAt
     */
    constructor(chatId, content, role, createdAt) {
        this.content = { chatId, content, role, createdAt };
        this.api = new Api();
    }

    createHTML = () => {
        this.htmlElement = document.createElement("div");
        const msgTextElem = document.createElement("div");
        msgTextElem.classList.add("text");
        msgTextElem.innerText = this.content.content;
        this.htmlElement.classList.add("message", this.content.role);
        this.htmlElement.appendChild(msgTextElem);
    };

    /** @param {string} text */
    addText = (text) => {
        this.content.content += text;
        if (this.htmlElement)
            this.htmlElement.firstElementChild.textContent += text;
    };

    /** @returns {Object} The JSON representation of the message object.*/
    formJson = () => {
        const { chat_id, content, role, created_at } = this.content;
        return { chat_id, content, role, created_at };
    };
}

export { Ollama, Api, Chat, Message, makeId, ollamaUrl };
