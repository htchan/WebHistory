import 'dart:convert';

import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:webhistory/models/all_models.dart';
import 'package:http/http.dart' as http;

class WebHistoryRepostory {
  String host, authToken;
  final String protocol = "http";
  final String api_prefix = dotenv.get("WEB_WATCHER_API_ROUTE_PREFIX", fallback: "/api/web-watcher");

  WebHistoryRepostory(this.host, this.authToken);

  String get url { return "${protocol}://${host}${api_prefix}"; }

  Future<List<WebGroup>> getWebGroups() {
    return http.get(Uri.parse("${this.url}/websites/groups"), headers: {'Authorization': this.authToken})
      .then((response) {
        Map<String, dynamic> responseMap = Map.from(jsonDecode(response.body));
        if (response.statusCode != 200) throw responseMap['message']??responseMap['error'];
        List responseGroups = List.from(responseMap['website_groups']);
        List<List<Web>> groups = List.from(responseGroups.map( (group) {
          return List.from(group).map( (item) {
            return Web.from(Map<String, String>.from(item));
          }).toList();
        }));
        return groups.map((group) => WebGroup(group)).toList();
      });
  }

  Future<WebGroup?> getWebGroup(String groupName) {
    return http.get(Uri.parse("${this.url}/websites/groups/${groupName}"), headers: {'Authorization': this.authToken})
      .then((response) {
        Map<String, dynamic> responseMap = Map.from(jsonDecode(response.body));
        if (response.statusCode != 200) throw responseMap['message']??responseMap['error'];
        List responseGroup = List.from(responseMap['website_group']);
        List<Web> group = List.from(responseGroup.map( (item) {
          return Web.from(Map<String, String>.from(item));
        }));
        return WebGroup(group);
      });
  }

  Future<Popup?> createWeb(String url) {
    return http.post(
      Uri.parse("${this.url}/websites/"),
      body: <String, String> {
        'url': url
      },
      headers: {'Authorization': this.authToken}
    )
    .then((response) {
      Map<String, String> responseMap = Map.from(jsonDecode(response.body));
      return Popup.from(responseMap);
    });
  }

  Future<Popup?> refreshWeb(String uuid) {
    print(uuid);
    return http.put(
      Uri.parse("${this.url}/websites/${uuid}/refresh"),
      headers: {'Authorization': this.authToken}
    )
    .then((response) {
      Map<String, dynamic> responseMap = Map.from(jsonDecode(response.body));
      return Popup.from(<String, String> {"message": "success"});
    });
  }

  Future<Popup?> chagneGroupName(String uuid, groupName) {
    return http.put(
      Uri.parse("${this.url}/websites/${uuid}/change-group"),
      body: <String, String> {
        "group_name": groupName
      },
      headers: {'Authorization': this.authToken})
      .then((response) {
        Map<String, String> responseMap = Map.from(jsonDecode(response.body));
        return Popup.from(responseMap);
      });
  }

  Future<Popup?> delete(String uuid) {
    return http.delete(Uri.parse("${this.url}/websites/${uuid}/"), headers: {'Authorization': this.authToken})
      .then((response) {
        Map<String, String> responseMap = Map.from(jsonDecode(response.body));
        return Popup.from(responseMap);
      });
  }
}