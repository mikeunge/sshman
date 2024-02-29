# SSHMan - SSH management done right

> Manage your ssh profiles with ease.

## About

SSHMan is a CLI tool for managing ssh keys & connection strings.

### But why?

We have all been there, typing ```ssh -i``` and pressing the UP arrow till our fingers hurt to find the right connection string for the server we want to connect to.

This tool is here to solve this.

### Commands

```bash
Available Options:

  -h  --help              Print help information.
  -v  --version           Prints the version.
      --about             Prints information about the program.
  -l  --list              List of all available SSH profiles.

  -c  --connect           Connect to a server with profile.
  -n  --new               Create a new SSH profile.
  -u  --update            Update an SSH profile.
  -d  --delete            Delete SSH profiles.
  -e  --export            Export profiles (for eg. sharing).

  -i  --id                Provide the id for directly using a profile.
  -a  --alias             Provide the alias for directly using a profile.
```
## Special thanks

- Thanks to [@atotto](https://gist.github.com/atotto/ba19155295d95c8d75881e145c751372) for this genius gist.

