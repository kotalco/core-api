#! /bin/bash
read -p "Type [seed] for seeding or Type [trunc] to truncate the db: " Answer
case $Answer in
 seed | SEED)
 go run pkg/cmd/seeder/main.go -a seed
  ;;
trunc)
  go run pkg/cmd/seeder/main.go -a trunc
  ;;

  *)
    echo "Invalid flag value"

esac