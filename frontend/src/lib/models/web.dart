import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:webhistory/repostories/webHistoryRepostory.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:intl/intl.dart';

class Web {
  final String uuid, url, title, groupName;
  final DateTime updateTime, accessTime;
  static final datetimeFormat = DateFormat(dotenv.get('WEB_DATETIME_FORMAT', fallback: 'yyyy-MM-ddThh:mm:ss zzz'));

  Web.from(Map<String, String> map):
    this.uuid = map['uuid']??"",
    this.url = map['url']??"",
    this.title = map['title']??"",
    this.updateTime = datetimeFormat.parse(map['update_time']??""),
    this.accessTime = datetimeFormat.parse(map['access_time']??""),
    this.groupName = map['group_name']??"";
  
  bool get isUpdated {
    return this.accessTime.isBefore(this.updateTime);
  }
  bool get isNull {
    return (uuid == "" && url == "" && title == "" && groupName == "");
  }

  Future<void> open(WebHistoryRepostory client) async {
        if (await canLaunch(url)) await launch(url);
        client.refreshWeb(uuid);
  }
}