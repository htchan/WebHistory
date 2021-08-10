import 'dart:convert';
// ignore: file_names
import 'package:flutter/material.dart';
// ignore: import_of_legacy_library_into_null_safe
import 'package:http/http.dart' as http;
import 'package:url_launcher/url_launcher.dart';
import '../Components/websiteCard.dart';

class MainPage extends StatefulWidget{
  final String url;

  const MainPage({Key? key, required this.url}) : super(key: key);

  @override
  _MainPageState createState() => _MainPageState(this.url);
}

class _MainPageState extends State<MainPage> {
  final String url;
  List<Widget> _web = [ const Center(child: Text("Loading")) ];
  // List<Widget> _buttons = _renderStageButton();
  final GlobalKey scaffoldKey = GlobalKey();

  _MainPageState(this.url) {
    _loadData();
  }

  void _loadData() {
    final String apiUrl = '$url/list';
    http.get(Uri.parse(apiUrl))
    .then((response) {
      if (response.statusCode >= 200 && response.statusCode < 300) {
          Map<String, dynamic> body = Map.from(jsonDecode(response.body));
          List websiteGroups = List<List>.from(body['websiteGroups']);
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

  @override
  Widget build(BuildContext context) {
    // show the content
    return Scaffold(
      appBar: AppBar(
        title: const Text('Web History'),
        actions: [
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