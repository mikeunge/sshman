# SSHMan - SSH connection manager

![Logo](https://github.com/mikeunge/sshman/blob/main/assets/logo.png?raw=true)

> Manage multiple SSH connections with the help of profiles.

## About

SSHMan is a CLI tool that helps you manage multiple ssh server / profiles with ease.

We have all been there, typing ```ssh -i``` and pressing the UP arrow till our fingers hurt to find the right connection string for the server we want to connect to.
And in the worst case we aren't in the directory where our ssh keys are stored.

This tool is here to solve this issue with managed profiles you can simply select and directly connect to the desired server.

## How it works

If you haven't already, create a profile with ```sshman --new```, this opens up an interface where you provide all the essential information for the ssh profile to work.

You can than either list all the available profiles with ```sshman --list``` or connect directly to the newly created profile with ```sshman --connect```.
When connecting with a private key, the private key gets generated and deleted automatically for you so you don't have to worry about nothing.

### Command overview

```bash
Available Options:

      --version      Prints the version.
      --about        Print information about the app.
    -l --list        Connect to a server with profile.
    -c --connect     Connect to a server with profile.
    -n --new         Create a new SSH profile.
      --no-encrypt   Don't encrypt the profile.
    -u --update      Update an SSH profile.
    -d --delete      Delete SSH profiles.
       --export      Export profiles.
       --import      Import profiles.
    -a --alias       Provide an alias to directly access.
    -i --id          Provide an id for directly accessing.
       --decrypt     Decrypt the profile. (used for export)

```
## Special thanks

- [@atotto](https://gist.github.com/atotto/ba19155295d95c8d75881e145c751372) for this genius gist
- [@donvito](https://gist.github.com/donvito/efb2c643b724cf6ff453da84985281f8) for helping with my lack of knowledge regarding encryption

