// ignore: file_names
import 'dart:convert';

import 'package:flutter/material.dart';
// ignore: import_of_legacy_library_into_null_safe
import 'package:http/http.dart' as http;
import 'package:webhistory/Clients/webHistoryClient.dart';
import 'package:webhistory/Components/websiteCard.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'package:webhistory/WebHistory/Models/webGroup.dart';

class DetailsPage extends StatefulWidget{
  final String groupName;
  WebHistoryClient client;

  DetailsPage({Key? key, required this.groupName, required this.client}) : super(key: key);

  @override
  _DetailsPageState createState() => _DetailsPageState(this.client, this.groupName);
}


class _DetailsPageState extends State<DetailsPage> {
  WebHistoryClient client;
  final GlobalKey<FormState> scaffoldKey = GlobalKey<FormState>();
  final String groupName;
  WebGroup? group;

  _DetailsPageState(this.client, this.groupName) {
    _loadData();
  }

  void handleNoMatchGroup(int n) {
    String errorMessage = 
      "No webiste Match group - ${group?.latestWeb.groupName}\nYou will back to Home Page in #{n} seconds";
    setState(() {
      group = null;
    });
    Future.delayed(Duration(seconds: n),
      () => Navigator.of(context).pop());
  }

  void _loadData() {
    client.webGroup(groupName)
    .then( (group) {
      setState(() { this.group = group; });
    })
    .catchError( (e) => {
      handleNoMatchGroup(3)
    });
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

  List<Widget> renderWebsites() {
    if (group == null) return [];
    return group!.webs.map(
      (web) => WebsiteCard(
        client: client,
        group: WebGroup([web]),
        updateList: _loadData,
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
        children: renderWebsites()
      )
    );
  }
}