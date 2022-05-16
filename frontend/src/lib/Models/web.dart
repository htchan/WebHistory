class Web {
  final String uuid, url, title, groupName;
  final DateTime updateTime, accessTime;

  Web.from(Map<String, String> map):
    this.uuid = map['uuid']??"",
    this.url = map['url']??"",
    this.title = map['title']??"",
    this.updateTime = DateTime.parse(map['update_time']??""),
    this.accessTime = DateTime.parse(map['access_time']??""),
    this.groupName = map['group_name']??"";
  
  bool get isUpdated {
    return this.accessTime.isBefore(this.updateTime);
  }
  bool get isNull {
    return (uuid == "" && url == "" && title == "" && groupName == "");
  }
}