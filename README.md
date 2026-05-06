# sship

Copy your SSH public key to a remote host with one command.

```
sship [user@]host
```

## How it works

1. Picks your SSH public key (~/.ssh/id_ed25519.pub, id_ecdsa.pub, or id_rsa.pub)
2. If no key exists, offers to generate one with `ssh-keygen`
3. Ships it to the remote host via `ssh`, creating ~/.ssh with the right permissions

Requires the system `ssh` and `ssh-keygen` commands (built into Windows 10+).

## Build

```sh
go build -o sship.exe .
```

Drop `sship.exe` anywhere in your PATH.
