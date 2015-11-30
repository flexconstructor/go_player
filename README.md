# GoPlayer
Pseudo-streaming live video player with golang
The application takes an rtmp stream from the camera and gives the user on web-socket stream of images in a format jpg.
To use this application you must install ffmpeg.
Unfortunately, at the moment, for various reasons it was not possible to achieve the stated functionality. 
Therefore, it was decided to limit ourselves to the following:
This library controls the flow of images generated ffmpeg within nginx-rtmp-module. Pictures of unix.socket take and give them to the client through the Web-socket.
Example configuration nginx:
```
# Pseudo video stream application.
    application ps {
      live on;
      allow publish 127.0.0.1;
      allow play all;

      exec_push ffmpeg
      -y -loglevel verbose
      -re -i rtmp://[source_rtmp_host]:[source_rtmp_port]/[source_app_name]/[source_stream_key]
      -vcodec libx264 -an -vsync 1
      -g 10 -vf scale=0:0
      -pix_fmt yuvj422p -q:v 1
      -f image2  -vcodec mjpeg -update 1
      unix://[path_to unix.socket file]/$name.sock
      2>>/[path/to/log/dir]/logs/ffmpeg-$name.log;
    }
```