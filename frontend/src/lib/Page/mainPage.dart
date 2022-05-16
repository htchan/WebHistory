import 'dart:convert';
import 'package:collection/collection.dart';
// ignore: file_names
import 'package:flutter/material.dart';
// ignore: import_of_legacy_library_into_null_safe
import 'package:http/http.dart' as http;
import 'package:url_launcher/url_launcher.dart';
import 'package:webhistory/repostories/webHistoryRepostory.dart';
import 'package:webhistory/WebHistory/Models/webGroup.dart';
import '../Components/websiteCard.dart';
import 'package:fluttertoast/fluttertoast.dart';

class MainPage extends StatefulWidget {
  WebHistoryRepostory client;

  MainPage({Key? key, required this.client}) : super(key: key);

  @override
  _MainPageState createState() => _MainPageState(this.client);
}

class _MainPageState extends State<MainPage> {
  WebHistoryRepostory client;
  List<WebGroup>? groups;
  final GlobalKey scaffoldKey = GlobalKey();

  _MainPageState(this.client) {
    _loadData();
  }

  bool isWebsiteUpdated(Map website) {
    return website['update_time'].compareTo(website['access_time']) > 0;
  }

  void _loadData() {
    print("update");
    client.getWebGroups().then((groups) {
      setState( () { this.groups = groups; });
    })
    .catchError((e) {
      //TODO: show popups for the error message
      setState( () { this.groups = []; });
      resultToast(e.toString());
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
  void openInsertPage() {
    Navigator.pushNamed(
      scaffoldKey.currentContext!,
      '/add'
    )
    .then( (value) => _loadData() );
  }
  void openAllUnreadComic() {
    if (groups == null) return;
    // loop website groups
    Future.wait(
      groups!.where( (group) => group.latestWeb.isUpdated)
      .mapIndexed( (i, group) async {
        if (await canLaunch(group.latestWeb.url)) await launch(group.latestWeb.url);
        // and update backend server of opened website
        client.refreshWeb(group.latestWeb.uuid);
      })
    )
    .then( (response) { _loadData(); });
  }

  Widget renderWebsiteCards() {
    if (groups == null) return Center(child: Text("loading"));
    if (groups!.length != 0) {
      return ListView.separated(
        separatorBuilder: (context, index) => const Divider(height: 10,),
        itemCount: groups!.length,
        itemBuilder: (context, index) => WebsiteCard(
          client: client,
          group: groups![index],
          updateList: _loadData,
        ),
      );
    } else {
      return Center(child: Text("failed to load"));
    }
  }

  @override
  Widget build(BuildContext context) {
    // show the content
    return Scaffold(
      appBar: AppBar(
        title: const Text('Web History'),
        actions: [
          IconButton(
            onPressed: openAllUnreadComic,
            icon: const Icon(Icons.open_in_browser_outlined)
          ),
          IconButton(
            onPressed: openInsertPage, 
            icon: const Icon(Icons.add_circle),
          )
        ],
      ),
      key: scaffoldKey,
      body: Container(
        child: renderWebsiteCards(),
        margin: const EdgeInsets.symmetric(horizontal: 5.0),
      ),
    );
  }
}
