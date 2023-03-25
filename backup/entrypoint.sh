tables=$(PGPASSWORD=$PSQL_PASSWORD psql -h $PSQL_HOST -p $PSQL_PORT -U $PSQL_USER -d $PSQL_NAME -t -c "Select table_name From information_schema.tables Where table_type='BASE TABLE' and table_schema='public'")

mkdir -p /backup/$(date '+%Y-%m-%d')
for table in $tables;
  do PGPASSWORD=$PSQL_PASSWORD pg_dump -h $PSQL_HOST -p $PSQL_PORT -U $PSQL_USER -d $PSQL_NAME -t $table > /backup/$(date '+%Y-%m-%d')/$table.sql;
done;
