class RightMenuComponent extends MenuComponent {
  css_height    = "100vh";
  css_width     = "130px";
  css_float     = "right";
  css_position  = "relative";
  css_zIndex    = "100";
  css_marginTop = "120px";

  buttons = [
    new MenuButtonLeftComponent("", "Login",
      [], [], [], () => layerManager.switchModeCallback()),
    new MenuButtonLeftComponent("", "Sign Up",
      [], [], [], () => layerManager.switchModeCallback()),
  ]
}

function mainServerClient() {

  bodyComponent.addChild(new RightMenuComponent());

}

mainServerClient(); // :)
