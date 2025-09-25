# ollama-chat

A simple chat application using Ollama for local LLM inference.

## What does this program do?

`ollama-chat` allows you to chat with large language models (LLMs) running locally via Ollama. You can send messages and receive AI-generated responses, just like chatting with popular AI chatbots, but all processing happens on your own machine—offering privacy, speed, and customization.

## Setup

1. **Install Ollama**

   First, make sure you have [Ollama](https://ollama.com/) installed and running on your machine.

2. **Clone this repository**

   ```bash
   git clone https://github.com/RakanMyHusbando/ollama-chat.git
   cd ollama-chat
   ```

3. **Install Go**

   You need Go installed to run the server.  
   Download and install Go from [golang.org](https://golang.org/dl/).

4. **Install dependencies**

   Use Go modules to install dependencies:

   ```bash
   go mod tidy
   ```
   
5. **Set enviroment variables**
   ```bash
   cp .env.example .env
   ```
   `HOST` default is `localhost`  

7. **Start the application**

   ```bash
   go run .
   ```

8. **Begin chatting!**

   Open your browser and go to the URL (set in .env) to start a conversation with your local LLM.

## Documentation

- **Configuration**: You might need to specify the model name in a config or settings file (see project files for details).
- **Usage**: Type your prompt in the chat window and press enter to get responses from your chosen LLM.
- **Extending**: You can modify supported models and UI as needed.

## License

See [`LICENSE`](/LICENSE) for details.
