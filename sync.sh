#! /usr/bin/env  bash

rsync  -arv -u  --exclude=".git"  .  wwwuser@182.92.152.61:/home/wwwuser/signaling 
