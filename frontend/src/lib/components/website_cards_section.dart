import 'package:flutter/material.dart';
import 'package:webhistory/components/website_card.dart';
import 'package:webhistory/models/web_group.dart';
import 'package:webhistory/repostories/webHistoryRepostory.dart';

class WebsiteCardsSection extends StatelessWidget {
  final List<WebGroup>? groups;
  final WebHistoryRepostory client;
  final Function updateList;

  WebsiteCardsSection({required this.groups, required this.client, required this.updateList});

  @override
  Widget build(BuildContext context) {
    if (groups == null) return Center(child: Text("loading"));
    if (groups!.length == 0) {
      return Center(child: Text("failed to load"));
    } else {
      return ListView.separated(
        separatorBuilder: (context, index) => const Divider(height: 10,),
        itemCount: groups!.length,
        itemBuilder: (context, index) => WebsiteCard(
          client: client,
          group: groups![index],
          updateList: updateList,
        ),
      );
    }
  }
}