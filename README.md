# serve

Serve is a tool that helps you serve Single Page Applications on your server.

## Usage
Create a file called serve.yaml in the working directory of the `serve` binary and do `./serve`.

## Features
1. Can serve SPA apps with 404 files defaulting to index.html for client side router to work properly.
2. Has ability to setup dynamic reverse proxies so we can route requests without CORS or any issues.
3. Reverse proxy returns the entire HTTP Headers and response as is.
4. We can proxy any HTTP method to the specified backend URL.

## License
MIT