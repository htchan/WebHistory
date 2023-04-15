import os, datetime

backup_dir = "/backup"
def dir_to_date(dir):
  return datetime.datetime.strptime( os.path.basename(dir), '%Y-%m-%d').date()

def get_schema_sql_path():
  schema_sql_paths = [
    dir_path
    for (dir_path, dir_names, file_names) in os.walk(backup_dir) 
    if 'schema.sql' in file_names
  ]
  schema_sql_paths.sort()


  schema_sql_path_groups = []
  current_date = datetime.datetime.today().date()

  for schema_sql_path in schema_sql_paths:
    if schema_sql_path_groups == []:
      schema_sql_path_groups.append([schema_sql_path])
      continue
    latest_processed_schema_sql_date = dir_to_date(schema_sql_path_groups[-1][-1])
    current_schema_sql_date = dir_to_date(schema_sql_path)
    if ((latest_processed_schema_sql_date.year == current_schema_sql_date.year and current_schema_sql_date.year < current_date.year - 1) or
        (latest_processed_schema_sql_date.month == current_schema_sql_date.month and current_schema_sql_date.month < current_date.month - 1)):
      schema_sql_path_groups[-1].append(schema_sql_path)
    else:
      schema_sql_path_groups.append([schema_sql_path])
    
  for i, group in enumerate(schema_sql_path_groups):
    for j, path in enumerate(group):
      schema_sql_path_groups[i][j] = f'{path}/schema.sql'

  return schema_sql_path_groups

def get_data_sql_path():
  data_sql_paths = [
    dir_path
    for (dir_path, dir_names, file_names) in os.walk(backup_dir) 
    if 'data.sql' in file_names
  ]
  data_sql_paths.sort()

  data_sql_path_groups = []
  current_date = datetime.datetime.today().date()

  for data_sql_path in data_sql_paths:
    if data_sql_path_groups == []:
      data_sql_path_groups.append([data_sql_path])
      continue
    latest_processed_data_sql_date = dir_to_date(data_sql_path_groups[-1][-1])
    current_data_sql_date = dir_to_date(data_sql_path)
    if ((latest_processed_data_sql_date.year == current_data_sql_date.year and current_data_sql_date.year < current_date.year - 1) or
        (latest_processed_data_sql_date.month == current_data_sql_date.month and current_data_sql_date.month < current_date.month - 1)):
      data_sql_path_groups[-1].append(data_sql_path)
    else:
      data_sql_path_groups.append([data_sql_path])
    
  for i, group in enumerate(data_sql_path_groups):
    for j, path in enumerate(group):
      data_sql_path_groups[i][j] = f'{path}/data.sql'

  return data_sql_path_groups

  return data_sql_paths

def _read_file(file_path):
  with open(file_path, 'r') as file:
    content = file.read()
    return content

def _write_file(file_path, content):
  with open(file_path, 'w') as file:
    file.write(content)

def process_schema_file(
  processed_schema_commands: list[str],
  reference_schema_sql: str,
  schema_file_path: str
):
  print(reference_schema_sql, schema_file_path)
  if reference_schema_sql != schema_file_path:
    _write_file(schema_file_path.replace("schema.sql", "reference_sql"), reference_schema_sql)
  schema_content = _read_file(schema_file_path)

  schema_commands = schema_content.split(';')
  schema_ending = schema_commands[-1]
  schema_commands = schema_commands[:-1]
  filtered_schema_commands = [ command for command in schema_commands if command not in processed_schema_commands ]
  filtered_schema_commands += [schema_ending]

  processed_schema_commands = processed_schema_commands[:-1] + filtered_schema_commands + processed_schema_commands[-1:]
  
  if len(filtered_schema_commands) == 1:
    os.remove(schema_file_path)
    return processed_schema_commands

  _write_file(schema_file_path, ';'.join(processed_schema_commands))
  
  return processed_schema_commands

def process_data_file(
  processed_data_commands: list[str],
  data_file_path: str
):
  data_content = _read_file(data_file_path)

  data_commands = data_content.split('\n\\.\n')
  data_ending = data_commands[-1]
  data_commands = data_commands[:-1]
  result_data_commands = []
  for command in data_commands:
    splitted_command = command.split(';\n')
    if len(splitted_command) == 1:
      continue

    key = splitted_command[0]
    actual_data = splitted_command[1].split('\n')
    processed_actual_data = processed_data_commands.get(key, [])

    filtered_actual_data = [ data for data in actual_data if data not in processed_actual_data ]
    if len(filtered_actual_data) == 0:
      continue

    result_data_commands += [';\n'.join([key, '\n'.join(filtered_actual_data)])]
    if len(processed_actual_data) == 0:
      processed_data_commands[key] = filtered_actual_data
    else:
      processed_data_commands[key] += filtered_actual_data

  result_data_commands += [data_ending]
  if len(result_data_commands) == 1:
    os.remove(data_file_path)
    return processed_data_commands
  
  _write_file(data_file_path, '\n\\.\n'.join(result_data_commands))
  
  return processed_data_commands

if __name__ == '__main__':
  schema_sql_groups = get_schema_sql_path()
  data_sql_groups = get_data_sql_path()

  print(schema_sql_groups)
  for schema_paths_group in schema_sql_groups:
    processed_schema_commands = []
    reference_schema_sql = schema_paths_group[0]
    for schema_sql_path in schema_paths_group:
      processed_schema_commands = process_schema_file(processed_schema_commands, reference_schema_sql, schema_sql_path)
      reference_schema_sql = schema_sql_path

  print(data_sql_groups)
  for data_paths_group in data_sql_groups:
    processed_data_commands = {}
    reference_data_sql = data_paths_group[0]
    for data_sql_path in data_paths_group:
      processed_data_commands = process_data_file(processed_data_commands, data_sql_path)
      reference_data_sql = data_sql_path

  ## TODO: only file within same time range should consider to turn into incremental backup
  ##       eg. in same week (non current week), in same month (not current month), in same year (not current year)
