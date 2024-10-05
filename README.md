# Key Excahnge
Key exchange is an educational implementation of the Wu-Lam key exchange protocol.

## Getting Started
To try the demonstration version, follow these steps:

Generate Keys: Run the key generation task by executing the following command:
```
task keygen-demo
```

Run Tasks: Open separate terminals and run the tasks for Trent, Alice, and Bob. For example, to run Alice's task, use:
```
task alice-run
```

To see all available tasks, use the following command:
```
task --list
```
## Usage
The tasks for Alice and Bob launch a Text User Interface (TUI). The first menu item allows the parties to generate a session key using the Wu-Lam protocol. After generating the key, Alice and Bob will be able to exchange messages securely.