## Add New Menu or Permission

Instructions :
1. Add json object at file `docs/menu-and-permission-list.json`
2. Run `make run-cron command="setup-predefined-role-menu-permission"`

### Add new menu (parent menu only)
1. Open json file at `docs/menu-and-permission-list.json`
2. Add json object at the bottom with a format like this :
```
{
    "name": "Home",
    "slug": "home",
    "icon": "pi-home",
    "path": "/dashboard",
    "type": "sidebar",
    "permissions": [
      {
        "slug": "home.view",
        "name": "View home",
        "group": "Home",
        "description": "View home",
        "roles": [
          "ADMIN",
          "DEVELOPER",
          "MAKER",
          "APPROVER",
          "OPERATION"
        ]
      }
    ]
  },
```
3. For new menus, new permissions must be created, at least for view permissions.
   If the menu page contains an action such as create, then also add the create permission in the `permissions` section.
4. The roles in the permissions are the default roles that can have these permissions.

### Add new menu (parent menu and child menu)
1. Open json file at `docs/menu-and-permission-list.json`
2. Add json object at the bottom with a format like this :
```
{
    "name": "Disbursement",
    "slug": "disbursement",
    "icon": "pi-send",
    "path": "/disbursement",
    "type": "sidebar",
    "permissions": [
      {
        "slug": "disbursement.view",
        "name": "View disbursement",
        "group": "Disbursement",
        "description": "View disbursement",
        "roles": [
          "ADMIN",
          "MAKER",
          "APPROVER",
          "OPERATION"
        ]
      }
    ],
    "children": [
      {
        "name": "Approval",
        "slug": "disbursement-approval",
        "icon": "-",
        "path": "/disbursement/approval",
        "permissions": [
          {
            "slug": "disbursement.approval.view",
            "name": "View disbursement approval",
            "group": "Disbursement",
            "description": "View disbursement approval",
            "roles": [
              "ADMIN",
              "MAKER",
              "APPROVER",
              "OPERATION"
            ]
          },
          {
            "slug": "disbursement.approval.action",
            "name": "Action disbursement approval",
            "group": "Disbursement",
            "description": "Action disbursement approval",
            "roles": [
              "APPROVER"
            ]
          }
        ]
      }
    ]
  }
```
3. For menus that have children, add a new object in the `children` section with the same value as the parent.

### Add new menu (child menu only)
1. Open json file at `docs/menu-and-permission-list.json`
2. Find the parent menu that you want to add a child menu to.
3. Add a json object at the bottom in the desired parent's `children` section with a format like this.
```
{
    "name": "Approval",
    "slug": "disbursement-approval",
    "icon": "-",
    "path": "/disbursement/approval",
    "permissions": [
      {
        "slug": "disbursement.approval.view",
        "name": "View disbursement approval",
        "group": "Disbursement",
        "description": "View disbursement approval",
        "roles": [
          "ADMIN",
          "MAKER",
          "APPROVER",
          "OPERATION"
        ]
      },
      {
        "slug": "disbursement.approval.action",
        "name": "Action disbursement approval",
        "group": "Disbursement",
        "description": "Action disbursement approval",
        "roles": [
          "APPROVER"
        ]
      }
    ]
}
```

### Add new permission at existing menu
1. Open json file at `docs/menu-and-permission-list.json`
2. Add permissions to the menu that will be used in the `permissions` section, in a format like this.
```
{
    "slug": "disbursement.approval.view",
    "name": "View disbursement approval",
    "group": "Disbursement",
    "description": "View disbursement approval",
    "roles": [
      "ADMIN",
      "MAKER",
      "APPROVER",
      "OPERATION"
    ]
}
```

