#!/bin/bash

cd $1

../amspackager -o $2/$4.json -d 8 -i $3/$4_h264-426x240-400.mp4 -i $3/$4_h264-640x360-800.mp4 -i $3/$4_h264-854x480-1600.mp4 -i $3/$4_h264-1280x720-3000.mp4 -i $3/$4_aac-128.mp4 -i $3/$4_sub_eng.vtt -l eng -i $3/$4_sub_fra.vtt -l fra $5

cd ..

