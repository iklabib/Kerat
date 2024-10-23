# How it works?
We have a compiler container that receive source code, compile them to executable binary, and run said executable in another container. No brainer.

# Motivation
I slapped gvisor to container and call it a sandbox. In nutshell, just like Go Playground does. 
