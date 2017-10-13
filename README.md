# PHN
Hide images within the noise of another image.

## Usage
`phn [hidden image] [source image]`

if 'source image' is left out a solid background image will be generated
with the colour supplied by the `c` flag.

### examples
#### Encoding
This will hide my github avatar in the noise of another picture

`phn -o="hidden.png" "https://avatars1.githubusercontent.com/u/16108486?v=4&s=460" "https://i.imgur.com/gdnKZt1.png"`

#### Decoding
To retrieve the hidden image use
`phn -decode -o="decoded.png" "hidden.png" "https://i.imgur.com/gdnKZt1.png"`



## Flags
| Flag     | Description                                                          |
|----------|----------------------------------------------------------------------|
| w        | width to resize both images to                                       |
| h        | height to resize both images to                                      |
| c        | background colour of generated image                                 |
| d        | colour depth of encoded image                                        |
| o        | output path of encoded image                                         |
| decode   | decode the given image                                               |
| estimate | decode by attempting to extract the image noise with a gaussian blur |
| help     | print help information                                               |