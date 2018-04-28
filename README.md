Proof of concept of Symfony Liip Imagine Bundle proxy to AmazonS3

It takes request and check for existence resource in redis. If resource exists it serve image respnse strait from stream, if not additionaly fires Symfony controller to generate image.

Don't laugh
