import 'package:flutter/material.dart';
import 'package:webhistory/components/website_cards_section.dart';
import 'package:webhistory/repostories/webHistoryRepostory.dart';
import 'package:webhistory/models/all_models.dart';

class MainScreen extends StatefulWidget {
  final WebHistoryRepostory client;

  const MainScreen({Key? key, required this.client}) : super(key: key);

  @override
  _MainScreenState createState() => _MainScreenState(this.client);
}

class _MainScreenState extends State<MainScreen> {
  WebHistoryRepostory client;
  List<WebGroup>? groups;
  final GlobalKey scaffoldKey = GlobalKey();

  _MainScreenState(this.client) {
    _loadData();
  }

  void _loadData() {
    client.getWebGroups()
    .then( (groups) => setState( () { this.groups = groups; }) )
    .catchError((e) {
      setState( () { this.groups = []; });
      Popup(e.toString()).show();
    });
  }

  void openAllUnreadComic() {
    if (groups == null) return;
    // loop website groups
    Future.wait(
      groups!
      .where( (group) => group.latestWeb.isUpdated )
      .map( (group) => group.latestWeb.open(client) )
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
            onPressed: () {
              Navigator.pushNamed(context, '/add')
              .then( (value) => _loadData() );
            }, 
            icon: const Icon(Icons.add_circle),
          )
        ],
      ),
      key: scaffoldKey,
      body: Container(
        child: WebsiteCardsSection(
          client: client,
          groups: groups,
          updateList: _loadData,
        ),
        margin: const EdgeInsets.symmetric(horizontal: 5.0),
      ),
    );
  }
}
