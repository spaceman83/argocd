#! /bin/bash

curl -s https://hot.forthemug.com/hot_mess_2#stats > /tmp/tmp.txt
MUGS=$(cat /tmp/tmp.txt | grep -B14 "completion time" | awk 'NR==1' | tr -d " " | tr -d ,)
GIN=$(cat /tmp/tmp.txt | grep -B9 "completion time" | awk 'NR==1' | tr -d " " | tr -d , )
TOTSTATION=$(cat /tmp/tmp.txt | grep -A3 "Total ports" | awk 'NR==4' | tr -d " " | tr -d ,)

echo "# HELP 2hot_2messy_mugs_delivered"
echo "# TYPE 2hot_2messy_mugs_delivered counter"
echo "2hot_2messy_mugs_delivered $MUGS"
echo "# HELP 2hot_2messy_gin_delivered"
echo "# TYPE 2hot_2messy_gin_delivered counter"
echo "2hot_2messy_gin_delivered $GIN"
echo "# HELP 2hot_2messy_totstation"
echo "# TYPE 2hot_2messy_totstation counter"
echo "2hot_2messy_totstation $TOTSTATION"