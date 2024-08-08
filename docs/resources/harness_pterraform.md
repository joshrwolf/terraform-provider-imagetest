---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "imagetest_harness_pterraform Resource - terraform-provider-imagetest"
subcategory: ""
description: |-
  A harness created from a generic terraform invocation.
---

# imagetest_harness_pterraform (Resource)

A harness created from a generic terraform invocation.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `inventory` (Attributes) The inventory this harness belongs to. This is received as a direct input from a data.imagetest_inventory data source. (see [below for nested schema](#nestedatt--inventory))
- `name` (String) The name of the harness. This must be unique within the scope of the provided inventory.
- `path` (String) The path to the terraform source directory.

### Optional

- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))
- `vars` (String) A json encoded string of variables to pass to the terraform invocation. This will be passed in as a .tfvars.json var file.

### Read-Only

- `id` (String) The unique identifier for the harness. This is generated from the inventory seed and harness name.

<a id="nestedatt--inventory"></a>
### Nested Schema for `inventory`

Required:

- `seed` (String)


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) The maximum time to wait for the k3s harness to be created.