for i in {1..10}; do
  curl -s "http://arbond.ru/go/14/notes/$i" > /dev/null &
done
time wait
