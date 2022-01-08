// ignore: file_names
import 'dart:convert';

import 'package:flutter/material.dart';
// ignore: import_of_legacy_library_into_null_safe
import 'package:http/http.dart' as http;
import 'package:webhistory/Components/websiteCard.dart';
import 'package:fluttertoast/fluttertoast.dart';

class DetailsPage extends StatefulWidget{
  final String url, groupName, token;

  const DetailsPage({Key? key, required this.url, required this.groupName, required this.token}) : super(key: key);

  @override
  _DetailsPageState createState() => _DetailsPageState(this.url, this.groupName, this.token);
}


class _DetailsPageState extends State<DetailsPage> {
  final String url, groupName, token;
  final GlobalKey<FormState> scaffoldKey = GlobalKey<FormState>();
  List<Widget> websiteGroup = [];

  _DetailsPageState(this.url, this.groupName, this.token) {
    _loadData();
  }

  void handleNoMatchGroup(int n) {
    String errorMessage = 
      "No webiste Match group - ${groupName}\nYou will back to Home Page in #{n} seconds";
    setState(() {
      websiteGroup = [ Center(child: Text(errorMessage)) ];
    });
    Future.delayed(Duration(seconds: n),
      () => Navigator.of(context).pop());
  }

  void _loadData() {
    final String apiUrl = '$url/list';
    http.get(Uri.parse(apiUrl), headers: {"Authorization": token})
    .then((response) {
      if (response.statusCode >= 200 && response.statusCode < 300) {
          Map<String, dynamic> body = Map.from(jsonDecode(response.body));
          List<Map<String, String>> targetWebsiteGroup = List<List>.from(body['websiteGroups'])
            .firstWhere(
              (websiteGroup) => Map<String, String>.from(websiteGroup[0])["groupName"] == groupName,
              orElse: () => []
            )
            .map( (item) => Map<String, String>.from(item) ).toList();
          if (targetWebsiteGroup.length > 0) {
            setState(() { websiteGroup = renderWebsites(targetWebsiteGroup); });
          } else {
            handleNoMatchGroup(3);
          }
          
      } else {
        handleNoMatchGroup(3);
      }
    });
  }

  Future<bool?> showDialog2(Map website) {
    TextEditingController groupNameText = TextEditingController();
    return showDialog<bool>(
      context: context, 
      builder: (context) => AlertDialog(
        title: Text("Change Group name"),
        content: TextFormField(
          controller: groupNameText,
          decoration: const InputDecoration(hintText: "Group Name"),
        ),//Text("Please input the new group name"),
        actions: [
          FlatButton(
            child: Text("Cancel"),
            onPressed: () => Navigator.of(context).pop(),
          ),
          FlatButton(
            child: Text("Change"),
            onPressed: () {
              final String apiUrl = '$url/group/change';
              // send the group name to server
              http.post(
                Uri.parse(apiUrl),
                body: <String, String> {
                  "url": website["url"],
                  "groupName": groupNameText.text
                },
                headers: {"Authorization": token}
              )
              .then( (response) {
                if (response.statusCode >= 200 && response.statusCode < 300) {
                  _loadData();
                  Navigator.of(context).pop();
                  return;
                }
                Map<String, String> data = Map<String, String>.from(jsonDecode(response.body));
                String msg = data["error"] ?? data["message"] ?? "Unknown Error";
                // show toast for result
                resultToast(msg);
              });
            },
          )
        ],
      )
    );
  }
  
  void resultToast(String msg) {
    Fluttertoast.showToast(
        msg: msg,
        toastLength: Toast.LENGTH_LONG,
        gravity: ToastGravity.BOTTOM,
        timeInSecForIosWeb: 5,
        fontSize: 16.0,
        backgroundColor: Colors.grey.shade300,
        textColor: Colors.black,
        webBgColor: "#DDDDDD",
        webPosition: "center",
    );
  }

  List<Widget> renderWebsites(List<Map<String, String>> list) {
    return list.map(
      (website) => WebsiteCard(
        url,
        website,
        this.token,
        _loadData,
        (dummy) => null,
        isEdit: true,
        showChangeGroupDialog: showDialog2,
      )
    ).toList();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Web Group Details'),
      ),
      body: Column(
        children: websiteGroup
      )
    );
  }
}