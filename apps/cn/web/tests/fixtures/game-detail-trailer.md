# Detail media fixture

`game-detail-trailer.webm` is an original two-second, silent, solid-color VP9
clip recorded from an original 320×180 canvas (`#394b44`, 10 fps, VP9 at 20 kbit/s)
using Chromium's MediaRecorder for the real HTMLVideoElement regression. No third-party media.
It is never a product asset or a visual golden.

An equivalent clip can be produced with FFmpeg (not a project/runtime dependency):

```sh
ffmpeg -f lavfi -i 'color=c=0x394b44:s=320x180:r=10' -t 2 -an -c:v libvpx-vp9 -b:v 20k game-detail-trailer.webm
```
