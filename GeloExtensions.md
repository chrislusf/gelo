GeloExtensions is a collection of useful but nonnecessary enhancements to GeloCore. It provides no commands and is merely a collection of utilities.

GeloExtensions defines:
  * ListBuilder - A utility type for creating [lists](GeloTypes#List.md) in [alien](GeloTypes#Alien.md) commands
  * ArgumentParser - a versatile and powerful parser for Gelo [lists](GeloTypes#List.md) that allows you to write very expressive commands.

It defines the following [ports](GeloTypes#Port.md):
  * Stdio and Stderr, that wrap Stdin/Stdout and Stderr, respectively.
  * `Tee(ps ...Port)`, a wrapper type that sends on all `ps` and recieves all `ps`.
  * `Couple(in, out Port)` creates a port that `Recv`'s from `in` and `Send`'s on `out`.