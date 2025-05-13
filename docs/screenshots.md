# Mockoho CLI Interface Screenshots

This document provides screenshots of the Mockoho CLI interface to help users understand how to use the application.

## Main Interface

The main interface of Mockoho consists of two panels:

1. **Features Panel** (left): Lists all available features (groups of endpoints)
2. **Endpoints Panel** (right): Lists all endpoints for the selected feature

```
┌─Mockoho - Server: Stopped | Proxy: https://api.real-server.com─────────────────────────────────────┐
│                                                                                                     │
├─Features───────────────────┬─Endpoints (users)──────────────────────────────────────────────────────┤
│                            │                                                                        │
│ > users                    │ > GET /api/users/:id 🟢                                                │
│   products                 │   [★standard | premium | error]                                        │
│   auth                     │                                                                        │
│                            │ > POST /api/users 🟢                                                   │
│                            │   [★success | validation-error]                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
│                            │                                                                        │
├────────────────────────────┴────────────────────────────────────────────────────────────────────────┤
│ t toggle  r response  o open  n new  d delete  p proxy  s server  q quit  h help                    │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Help Dialog

Pressing `h` displays the help dialog with all available keyboard shortcuts:

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                                     │
│                                          Mockoho Help                                               │
│                                                                                                     │
│ Navigation:                                                                                         │
│   Tab       - Switch between Features and Endpoints panels                                          │
│   ↑/↓       - Navigate up/down in the current panel                                                │
│   Enter     - Select a feature or endpoint                                                          │
│                                                                                                     │
│ Actions:                                                                                            │
│   t         - Toggle endpoint active/inactive                                                       │
│   r         - Cycle through available responses (sets as default)                                   │
│   o         - Open configuration file in default editor                                             │
│   n         - Create new endpoint or feature                                                        │
│   d         - Delete selected endpoint or feature                                                   │
│   p         - Change proxy target                                                                   │
│   s         - Start/stop server                                                                     │
│   q         - Quit application                                                                      │
│   h         - Show this help screen                                                                 │
│   /         - Search for endpoints                                                                  │
│   Ctrl+r    - Reload configurations from disk                                                       │
│                                                                                                     │
│ Press Esc or any key to return...                                                                   │
│                                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## New Feature Dialog

Pressing `n` in the Features panel displays the new feature dialog:

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                                     │
│                                      Create New Feature                                             │
│                                                                                                     │
│ ▎                                                                                                   │
│ Feature name                                                                                        │
│                                                                                                     │
│                                                                                                     │
│ [Enter] Confirm  [Esc] Cancel                                                                       │
│                                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## New Endpoint Dialog

Pressing `n` in the Endpoints panel displays the new endpoint dialog:

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                                     │
│                                      Create New Endpoint                                            │
│                                                                                                     │
│ ▎                                                                                                   │
│ Endpoint ID                                                                                         │
│                                                                                                     │
│                                                                                                     │
│                                                                                                     │
│ Method (GET, POST, PUT, DELETE)                                                                     │
│                                                                                                     │
│                                                                                                     │
│                                                                                                     │
│ Path (e.g., /api/users/:id)                                                                         │
│                                                                                                     │
│                                                                                                     │
│ [Enter] Confirm  [Esc] Cancel                                                                       │
│                                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Delete Confirmation Dialog

Pressing `d` displays the delete confirmation dialog:

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                                     │
│                                        Confirm Delete                                               │
│                                                                                                     │
│ Are you sure you want to delete endpoint 'get-user-profile'?                                        │
│                                                                                                     │
│ [Enter] Confirm  [Esc] Cancel                                                                       │
│                                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Proxy Configuration Dialog

Pressing `p` displays the proxy configuration dialog:

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                                     │
│                                     Proxy Configuration                                             │
│                                                                                                     │
│ https://api.real-server.com▎                                                                        │
│ Proxy target URL                                                                                    │
│                                                                                                     │
│                                                                                                     │
│ [Enter] Confirm  [Esc] Cancel                                                                       │
│                                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Server Running

When the server is running, the header shows the server status and address:

```
┌─Mockoho - Server: Running (localhost:3000) | Proxy: https://api.real-server.com───────────────────┐
│                                                                                                     │
├─Features───────────────────┬─Endpoints (users)──────────────────────────────────────────────────────┤
```

## Toggling Endpoints

Pressing `t` toggles an endpoint between active (🟢) and inactive (🔴):

```
│ > GET /api/users/:id 🟢                                                │
│   [★standard | premium | error]                                        │
```

```
│ > GET /api/users/:id 🔴                                                │
│   [★standard | premium | error]                                        │
```

## Cycling Responses

Pressing `r` cycles through available responses for an endpoint:

```
│ > GET /api/users/:id 🟢                                                │
│   [★standard | premium | error]                                        │
```

```
│ > GET /api/users/:id 🟢                                                │
│   [standard | ★premium | error]                                        │
```

```
│ > GET /api/users/:id 🟢                                                │
│   [standard | premium | ★error]                                        │
```

## Search

Pressing `/` activates the search mode:

```
┌─Mockoho - Server: Running (localhost:3000) | Proxy: https://api.real-server.com───────────────────┐
│                                                                                                     │
├─Features───────────────────┬─Endpoints (users)──────────────────────────────────────────────────────┤
│ Search: user▎              │                                                                        │
```

These ASCII representations provide a visual guide to the Mockoho CLI interface. In a real application, these would be actual screenshots of the application running in a terminal.
