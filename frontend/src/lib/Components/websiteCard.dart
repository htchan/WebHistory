import 'dart:convert';
import 'dart:js';
// ignore: file_names
import 'package:flutter/material.dart';
// ignore: import_of_legacy_library_into_null_safe
import 'package:http/http.dart' as http;
import 'package:url_launcher/url_launcher.dart';
import 'package:flutter_slidable/flutter_slidable.dart';
import 'package:webhistory/repostories/webHistoryRepostory.dart';
import 'package:webhistory/Components/statusButton.dart';
import 'package:webhistory/Models/webGroup.dart';
import 'package:fluttertoast/fluttertoast.dart';

class WebsiteCard extends StatelessWidget {
  final WebHistoryRepostory client;
  final WebGroup group;
  final Function updateList;
  // final String token;
  const WebsiteCard({
    required this.client,
    required this.group,
    required this.updateList,
  });

  void openURL() async {
    // refresh web and call update list
    client.refreshWeb(group.latestWeb.uuid)
    .then( (response) { updateList(); } );
    // TODO: if it is not available to launch, it have to give a pop up
    if (await canLaunch(group.latestWeb.url)) await launch(group.latestWeb.url);
  }

  Text renderSubTitleText() {
    return Text(
      (group.latestWeb.url) + '\n' +
      'Update Time: ' + group.latestWeb.updateTime.toLocal().toString() + '\n' +
      'Access Time: ' + group.latestWeb.accessTime.toLocal().toString()
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
          FlatButton(
            child: Text("Cancel"),
            onPressed: () => Navigator.of(context).pop(),
          ),
          FlatButton(
            child: Text("Change"),
            onPressed: () {
              client.chagneGroupName(group.latestWeb.uuid, groupNameText.text)
              .then((group) { Navigator.of(context).pop(); })
              .catchError((e) { resultToast(e); });
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
            '/details?groupName=${group.latestWeb.groupName}'
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
          leading: WebsiteCardStatusButton(checked: group.latestWeb.isUpdated),
          title: Text(group.latestWeb.groupName),
          subtitle: renderSubTitleText(),
        ),
      ),
      actions: renderActions(context),
      secondaryActions: renderActions(context),
    );
  }
}