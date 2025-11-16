# Layer — Real-Time Collaborative Notes

A high-performance backend for real-time collaborative note-taking, built with Go, PostgreSQL, and WebSockets.  
Includes authentication, note management, collaborator permissions, and live editing synchronization.

## Overview

Layer is a real-time collaborative notes backend built with Golang.  
It provides secure authentication, note management, collaborator permissions, and WebSocket-based live editing.  
The system is optimized for low-latency synchronization and scalable multi-user collaboration.

## Features

- User authentication with access & refresh tokens
- Secure password hashing (bcrypt)
- Notes CRUD with ownership rules
- Collaborator system with access control
- Real-time editing over WebSockets
- Presence updates for connected users
- Automatic state initialization on connect
- Patch broadcasting to all clients in a note room
- Persistent content updates to PostgreSQL

## Tech Stack

- **Golang** — primary backend language
- **Gorilla Mux** — HTTP router
- **PostgreSQL** — primary database
- **golang-migrate** — database migrations
- **bcrypt** — password hashing
- **JWT (HS256)** — authentication tokens
- **WebSockets** — real-time collaboration
- **Validator v10** — payload validation

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/leugard21/layer-api.git
   cd layer-api
   ```
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Apply database migrations:
   ```bash
   make migrate-up
   ```
4. Start the server:
   ```bash
   make run
   ```
