# Project Name

A brief description of the project.

## Table of Contents

- [Project Name](#project-name)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
    - [Prerequisites](#prerequisites)
    - [Installation Steps](#installation-steps)
  - [Usage](#usage)
    - [Examples plus how to use](#examples-plus-how-to-use)
      - [Go Installation](#go-installation)
      - [Project Setup](#project-setup)
  - [Releases](#releases)
    - [Release 1.0](#release-10)
    - [Release 2.0](#release-20)

## Installation

You can run the project using the release for your systems platform.

### Prerequisites

You need to have go installed on your computer to run this project 
But with the executable you can just run it.

### Installation Steps

Get the release for your platform 

Unzip the zip file containing the binary for your platform 

Then cd into the extracted directory and run the following in your terminal
´´´Bash
./chitty-chat
´´´

## Usage

When the program first starts it will ask you what you want to do
is start a Server as shown below 

´´´
? Select Action: 
  ▸ Start a New Server
    Create a New Client
´´´
You can only start one server 

And then you can run the program in a new terminal to start a client to connect to the server 
´´´
? Select Action: 
    Start a New Server
  ▸ Create a New Client
´´´
You can create multiple clients 
When you start the client you will get prompted to give the client a name and port

´´´
✔ Give a name for the client or leave empty for default settings: █
✔ Give a port for the client or leave empty for default settings: 
´´´



### Examples plus how to use

you can leave these blank which would create a client with  

´´´
address = "localhost:4321"
name = "chit-chat"
´´´

You cannot make clients with the same name.

After your client has joined the server you will be able to chat or leave the chat server 

´´´
Use the arrow keys to navigate: ↓ ↑ → ← 
? Select Action: 
  ▸ Leave
    chat
 
´´´
if you choice chat you will be prompted to send a message.

´´´
✗ input your message and send : █
´´´

and then you kan send by clicking enter
and when you are done you will be guided back to action menu 

´´´
Use the arrow keys to navigate: ↓ ↑ → ← 
? Select Action: 
  ▸ Leave
    chat
 
´´´
#### Go Installation

1. Download the Go installation package from the official website.
2. Follow the installation instructions for your operating system.

#### Project Setup

1. Clone the repository.
2. Run `go mod download` to download the project dependencies.
3. Run `go build` to build the project.

## Releases

### Release 1.0

- server and client are in a sperate file

### Release 2.0

- now you have a startup ui
