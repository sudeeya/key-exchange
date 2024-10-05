# Key Excahnge
Key exchange is an educational implementation of the Wu-Lam key exchange protocol. The Wu-Lam protocol is a cryptographic protocol designed to securely exchange keys between two parties with the help of a trusted third party (Trent).

## Getting Started
To try the demonstration version, follow these steps:

1. Generate Keys. Run the RSA key generation task by executing the following command:
```
task keygen-demo
```
These keys will be required for message exchange in the Wu-Lam protocol.

2. Run Tasks. Open separate terminals and run the tasks for Trent, Alice, and Bob. For example, to run Alice's task, use:
```
task alice-run
```

To see all available tasks, use the following command:
```
task --list
```

## Usage
The tasks for Alice and Bob launch a Text User Interface (TUI).

The first menu item allows the parties to generate a session key using the Wu-Lam protocol. Select item by pressing Enter and write ID of your interlocutor (Alice's ID is "alice", Bob's is "bob"). It is enough to do this action on one side.

After generating the key, Alice and Bob will be able to exchange messages securely.