# go-firebird

## Installation
### Prerequisites:
- Go 
- Docker 
- Git 

### Clone the Repository:
Run the following command to clone the project:
`git clone https://github.com/brayanMuniz/go-firebird.git`
Then navigate into the project folder:
`cd go-firebird`

## Setting Up Development Environment

### Install Dependencies:
Ensure you have all the required dependencies installed:
`go mod tidy`

### Environment Variables:
Create a `.env` file in the project root and add the required API keys:

OPENAI_API_KEY=your_openai_api_key 

## Running with Air for Hot Reloading
Air is used for live reloading during development

### Install Air:
Run the following command to install Air globally:
`go install github.com/cosmtrek/air@latest`
Ensure your `$GOPATH/bin` is added to the system's PATH.

#### **MacOS:**
After installation, ensure that Air is in your PATH by adding the following to your `.zshrc` (or `.bashrc` if using Bash):  
`export PATH=$PATH:$(go env GOPATH)/bin`  

Then reload your terminal:  
`source ~/.zshrc`  

#### **Windows:**
On Windows, install Air using:  
`go install github.com/cosmtrek/air@latest`  

If the command is not recognized, ensure `GOPATH/bin` is added to the system's PATH:  
1. Open **Control Panel** → **System** → **Advanced system settings**.  
2. Click on **Environment Variables**.  
3. Edit the **Path** variable and add:  
   `%USERPROFILE%\go\bin`  
4. Restart your terminal.  

#### **Linux:**
Install Air using:  
`go install github.com/cosmtrek/air@latest`  

Ensure `GOPATH/bin` is in your PATH:  
`export PATH=$PATH:$(go env GOPATH)/bin`  

Reload the terminal:  
`source ~/.bashrc` (or `source ~/.zshrc` if using Zsh).  

> **Note:** If you encounter issues running `air`, verify that `$GOPATH/bin` is correctly set and that Go is properly installed.

### Initialize Air:
Run:
`air init`
This will generate an `.air.toml` configuration file in the project root.

### Start Development Server with Air:
To start the server with hot-reloading:
`air`

> If you don't want to use air and want to just start coding run `go run main.go` 

