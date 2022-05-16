import 'package:flutter/material.dart';

class WebsiteCardStatusButton extends StatelessWidget {
  final bool checked;

  WebsiteCardStatusButton({required this.checked});

  @override
  Widget build(BuildContext context) {
    if (checked) {
      return Container(
        color: Colors.green,
        child: const Icon(Icons.check_circle, color: Colors.white)
      );
    } else {
      return Container(
        color: Colors.red,
        child: const Icon(Icons.remove_circle, color: Colors.white)
      );
    }
  }
}