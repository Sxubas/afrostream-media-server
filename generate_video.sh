#!/bin/bash

dir_in=$1
dir_out=$2
video_name=$3

if [[ $# -gt 3 ]]; then
    video_ext=$4
else
    video_ext=mp4
fi

ffmpeg -i ${dir_in}/${video_name}.${video_ext} -y -vf yadif=0:-1:0,scale="426:trunc(ow/a/2)*2",setsar=1:1 -c:v libx264 -preset:v fast -profile:v baseline -level 3.0 -coder:v 0 -b:v 400k -minrate 400k -maxrate 400k -bufsize 400k -g 50 -keyint_min 50 -sc_threshold 0 -pix_fmt yuv420p -map 0:v -an -map_chapters -1 -threads 0 ${dir_out}/${video_name}_h264-426x240-400.mp4
ffmpeg -i ${dir_in}/${video_name}.${video_ext} -y -vf yadif=0:-1:0,scale="640:trunc(ow/a/2)*2",setsar=1:1 -c:v libx264 -preset:v fast -profile:v baseline -level 3.0 -coder:v 0 -b:v 800k -minrate 800k -maxrate 800k -bufsize 800k -g 50 -keyint_min 50 -sc_threshold 0 -pix_fmt yuv420p -map 0:v -an -map_chapters -1 -threads 0 ${dir_out}/${video_name}_h264-640x360-800.mp4
ffmpeg -i ${dir_in}/${video_name}.${video_ext} -y -vf yadif=0:-1:0,scale="854:trunc(ow/a/2)*2",setsar=1:1 -c:v libx264 -preset:v fast -profile:v baseline -level 3.1 -coder:v 1 -b:v 1600k -minrate 1600k -maxrate 1600k -bufsize 1600k -g 50 -keyint_min 50 -sc_threshold 0 -pix_fmt yuv420p -map 0:v -an -map_chapters -1 -threads 0 ${dir_out}/${video_name}_h264-854x480-1600.mp4
ffmpeg -i ${dir_in}/${video_name}.${video_ext} -y -vf yadif=0:-1:0,scale="1280:trunc(ow/a/2)*2",setsar=1:1 -c:v libx264 -preset:v fast -profile:v main -level 3.1 -coder:v 1 -b:v 3000k -minrate 3000k -maxrate 3000k -bufsize 3000k -g 50 -keyint_min 50 -sc_threshold 0 -pix_fmt yuv420p -map 0:v -an -map_chapters -1 -threads 0 ${dir_out}/${video_name}_h264-1280x720-3000.mp4
ffmpeg -i ${dir_in}/${video_name}.${video_ext} -y -vn -acodec aac -strict -2 -ac 2 -ab 128k -ar 48000 -map 0:a:0 -map_chapters -1 -threads 0 ${dir_out}/${video_name}_aac-128.mp4
