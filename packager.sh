#!/bin/bash

cd $1

../amspackager -o $2/$3.json -d 8 -i $3/$2_h264-426x240-400.mp4 -i $3/$2_h264-640x360-800.mp4 -i $3/$2_h264-854x480-1600.mp4 -i $3/$2_h264-1280x720-3000.mp4 -i $3/$2_aac-128.mp4 -i $3/$2_sub_fra -l fra $5

cd ..

