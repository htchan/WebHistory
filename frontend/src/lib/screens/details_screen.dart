import 'package:flutter/material.dart';
import 'package:webhistory/components/website_cards_section.dart';
import 'package:webhistory/models/popup.dart';
import 'package:webhistory/repostories/webHistoryRepostory.dart';
import 'package:webhistory/models/web_group.dart';

class DetailsScreen extends StatefulWidget{
  final String groupName;
  final WebHistoryRepostory client;

  const DetailsScreen({Key? key, required this.groupName, required this.client}) : super(key: key);

  @override
  _DetailsScreenState createState() => _DetailsScreenState(this.client, this.groupName);
}


class _DetailsScreenState extends State<DetailsScreen> {
  WebHistoryRepostory client;
  final GlobalKey<FormState> scaffoldKey = GlobalKey<FormState>();
  final String groupName;
  WebGroup? group;

  _DetailsScreenState(this.client, this.groupName) {
    _loadData();
  }

  void handleNoMatchGroup(int n) {
    Popup(
      "No webiste Match group - ${group?.latestWeb.groupName}\nYou will back to Home Screen in #{n} seconds"
    ).show();
    setState(() {
      group = null;
    });
    Future.delayed(Duration(seconds: n),
      () => Navigator.of(context).pop());
  }

  void _loadData() {
    client.getWebGroup(groupName)
    .then( (group) {
      setState(() { this.group = group; });
    })
    .catchError( (e) {
      handleNoMatchGroup(3);
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Web Group Details'),
      ),
      body: WebsiteCardsSection(
        client: client,
        groups: group?.webs.map( (web) => WebGroup([web]) ).toList()??[],
        updateList: _loadData,
      ),
    );
  }
}