# tinygo-wasm-webgl-demo

This repository contains a modified gowebapi/webapi [demo](https://github.com/gowebapi/webapi/blob/41cedfc27a0bd35c1220dd0fe4b4c4505c33b0ea/graphics/webgl/example_cube_test.go) that works with tinygo.

![demo](.github/demo.mov)

# Prerequisites
- Install [tinygo](https://tinygo.org/)
- Install [http-server](https://github.com/http-party/http-server) (or any other static web server of your choice). This is required to host the wasm file.

# Compilation
```bash
$ tinygo build -o=main.wasm -target=wasm -no-debug ./main.go

184K main.wasm
```

# Running
```bash
$ http-server .
```

# Resources:
- https://tinygo.org/ 
- https://github.com/gowebapi/webapi
- https://github.com/http-party/http-server
