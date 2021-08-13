const statusHelper = {
    isHomePage(path) {
        return path === "/home";
    },
    isSharePage(path) {
        return path && path.startsWith("/s/");
    },
    isAdminPage(path) {
        return path && path.startsWith("/admin");
    },
    isLoginPage(path) {
        return path && path.startsWith("/login");
    },
    isMobile() {
        return window.innerWidth < 600;
    },
};
export default statusHelper;
