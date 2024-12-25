const ollama = require("ollama");

const resp = await ollama.chat({
    model: "codellama:13b",
    message: [
        {
            role: "user",
            content: "Hello, how are you?",
        },
    ],
});

console.log(resp.message.content);
