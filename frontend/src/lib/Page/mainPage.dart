import 'dart:convert';
// ignore: file_names
import 'package:flutter/material.dart';
// ignore: import_of_legacy_library_into_null_safe
import 'package:http/http.dart' as http;
import 'package:url_launcher/url_launcher.dart';
import '../Components/websiteCard.dart';

class MainPage extends StatefulWidget {
  final String url;
  final String token;

  const MainPage({Key? key, required this.url, required this.token}) : super(key: key);

  @override
  _MainPageState createState() => _MainPageState(this.url, this.token);
}

class _MainPageState extends State<MainPage> {
  final String url;
  final String token;
  List<List> websiteGroups = [];
  List<Widget> _web = [ const Center(child: Text("Loading")) ];
  // List<Widget> _buttons = _renderStageButton();
  final GlobalKey scaffoldKey = GlobalKey();

  _MainPageState(this.url, this.token) {
    _loadData();
  }

  bool isWebsiteUpdated(Map website) {
    return website['updateTime'].compareTo(website['accessTime']) > 0;
  }

  void _loadData() {
    final String apiUrl = '$url/list';
    http.get(Uri.parse(apiUrl), headers: {"Authroization": token})
    .then((response) {
      if (response.statusCode >= 200 && response.statusCode < 300) {
          Map<String, dynamic> body = Map.from(jsonDecode(response.body));
          websiteGroups = List<List>.from(body['websiteGroups']);
          websiteGroups = [
            ...websiteGroups.where( (websiteGroup) => 
              websiteGroup.length > 0 ? isWebsiteUpdated(websiteGroup[0]) : false
            ),
            ...websiteGroups.where( (websiteGroup) =>
              websiteGroup.length > 0 ? !isWebsiteUpdated(websiteGroup[0]) : true
            ),
          ];
          print(websiteGroups.length);
          setState(() {
            _web = websiteGroups.map(
              (websiteGroup) {
                Map<String, String> website = Map<String, String>.from(websiteGroup[0]);
                website["title"] = website["groupName"]??"unknown";
                return WebsiteCard(url, website, _loadData, openDetailsPage);
              }
            ).toList();
          });
      } else {
        _web = [ const Center(child: Text("Failed to load data")) ];
      }
    });
  }
  
  void openDetailsPage(String groupName) {
    Navigator.pushNamed(
      scaffoldKey.currentContext!,
      '/details?groupName=${groupName}'
    )
    .then( (value) => _loadData() );
  }
  void openInsertPage() {
    Navigator.pushNamed(
      scaffoldKey.currentContext!,
      '/add'
    )
    .then( (value) => _loadData() );
  }
  void openAllUnreadComic() {
    print(websiteGroups[0]);
    print(isWebsiteUpdated(websiteGroups[0][0]));
    // loop website groups
    Future.wait(
      websiteGroups.where( (websiteGroup) {
        return isWebsiteUpdated(websiteGroup[0]);
      })
      .map( (websiteGroup) async {
        await canLaunch(websiteGroup[0]['url'])? await launch(websiteGroup[0]['url']) : "";
        // and update backend server of opened website
        final String apiUrl = '$url/websites/refresh';
        return http.post(
          Uri.parse(apiUrl),
          body: <String, String>{
            'url': websiteGroup[0]['url']??"",
          },
          headers: {"Authorization": token}
        );
      })
    )
    .then( (response) { _loadData(); });
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
        child: ListView.separated(
          separatorBuilder: (context, index) => const Divider(height: 10,),
          itemCount: _web.length,
          itemBuilder: (context, index) => _web[index],
        ),
        margin: const EdgeInsets.symmetric(horizontal: 5.0),
      ),
    );
  }
}