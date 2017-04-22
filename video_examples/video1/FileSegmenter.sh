ffmpeg -i sample.mp4 -c copy -map 0 -vbsf h264_mp4toannexb -f segment -segment_time 10 -segment_list test.m3u8 -segment_format mpegts segment%d.ts
