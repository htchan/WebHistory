// ignore: file_names
import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:flutter_slidable/flutter_slidable.dart';
import 'package:webhistory/repostories/webHistoryRepostory.dart';
import 'package:webhistory/models/all_models.dart';

import 'status_button.dart';

class WebsiteCard extends StatelessWidget {
  final WebHistoryRepostory client;
  final WebGroup group;
  final Web web;
  final Function updateList;
  static final Duration timeZoneOffset = DateTime.now().timeZoneOffset;
  // final String token;
  WebsiteCard({
    required this.client,
    required this.group,
    required this.updateList,
  }): this.web = group.latestWeb;

  void openURL() async {
    // refresh web and call update list
    client.refreshWeb(web.uuid)
    .then( (response) { updateList(); } )
    .catchError( (e) { Popup(e.toString()).show(); } );

    // TODO: if it is not available to launch, it have to give a pop up
    if (await canLaunch(web.url)) await launch(web.url);
  }

  Text renderSubTitleText() {
    return Text(
      (web.url) + '\n' +
      'Update Time: ' + web.updateTime.add(timeZoneOffset).toString() + '\n' +
      'Access Time: ' + web.accessTime.add(timeZoneOffset).toString()
    );
  }

  Future<bool?> showChangeGroupDialog(BuildContext context) {
    TextEditingController groupNameText = TextEditingController();
    return showDialog<bool>(
      context: context, 
      builder: (context) => AlertDialog(
        title: Text("Change Group name"),
        content: TextFormField(
          controller: groupNameText,
          decoration: const InputDecoration(hintText: "Group Name"),
        ),
        actions: [
          TextButton(
            child: Text("Cancel"),
            onPressed: () => Navigator.of(context).pop(),
          ),
          TextButton(
            child: Text("Change"),
            onPressed: () {
              client.chagneGroupName(web.uuid, groupNameText.text)
              .then((group) { Navigator.of(context).pop(); })
              .catchError((e) { Popup(e).show(); });
            },
          )
        ],
      )
    );
  }
  List<Widget> renderActions(BuildContext context) {
    IconSlideAction action;
    if (group.webs.length > 1) {
      // details page will be shown when there are more than one web in group
      action = IconSlideAction(
        caption: "Details",
        color: Colors.blue,
        icon: Icons.info,
        onTap: () {
          Navigator.pushNamed(
            context,
            '/details?groupName=${web.groupName}'
          )
          .then( (value) => updateList() );
        }
      );
    } else {
      // change group name popup will be shown when there are one web in group
      action = IconSlideAction(
        caption: "Change Group",
        color: Colors.yellow,
        icon: Icons.edit,
        onTap: () {
          print("working");
          // // show a dialog for input / select new group name (default group is user )
          showChangeGroupDialog(context);
          // // update the page
          updateList();
        }
      );
    }
    return [
      action,
      IconSlideAction(
        caption: 'Delete',
        color: Colors.red,
        icon: Icons.delete,
        onTap: () {
          group.webs.forEach((web) { client.delete(web.uuid); });
          updateList();
        }
      )
    ];
  }

  @override
  Widget build(BuildContext context) {
    return Slidable(
      actionPane: SlidableDrawerActionPane(),
      actionExtentRatio:0.2,
      child: GestureDetector(
        onTap: openURL,
        child:ListTile(
          leading: WebsiteCardStatusButton(checked: web.isUpdated),
          title: Text(web.groupName),
          subtitle: renderSubTitleText(),
        ),
      ),
      actions: renderActions(context),
      secondaryActions: renderActions(context),
    );
  }
}