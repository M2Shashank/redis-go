[![progress-banner](https://backend.codecrafters.io/progress/redis/77c8d335-e22c-4e2c-8cd4-d23f99c94813)](https://app.codecrafters.io/users/M2Shashank?r=2qF)

This is a starting point for Go solutions to the
["Build Your Own Redis" Challenge](https://codecrafters.io/challenges/redis).

In this challenge, you'll build a toy Redis clone that's capable of handling
basic commands like `PING`, `SET` and `GET`. Along the way we'll learn about
event loops, the Redis protocol and more.

**Note**: If you're viewing this repo on GitHub, head over to
[codecrafters.io](https://codecrafters.io) to try the challenge.

# Passing the first stage

The entry point for your Redis implementation is in `app/main.go`. Study and
uncomment the relevant code, then run the command below to execute the tests on
our servers:

```sh
codecrafters submit
```

That's all!

# Stage 2 & beyond

Note: This section is for stages 2 and beyond.

1. Ensure you have `go (1.27)` installed locally
1. Run `./your_program.sh` to run your Redis server, which is implemented in
   `app/main.go`.
1. Run `codecrafters submit` to submit your solution to CodeCrafters. Test
   output will be streamed to your terminal.


# Notes for Myself
Things to Modify

1. set with expiry has bugs. New set, invalidates earlier
2. BLPOP connection channels need to be popped of the map if clients disconnects
6. Validate the command before queueing - Transactions
9. remove the strong coupling of server and client somehow. Is it required ?
10. Rdb file contents can be restructured, seperate expiry keys from others and handle in store. 
 Parser can be moved to resp as a result ig
11. 