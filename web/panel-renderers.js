(function () {
  function renderExternalPanels(game, els, helpers) {
    if (typeof window.renderMiniGames === "function") {
      window.renderMiniGames(game, els.playdesk);
    }
    if (typeof window.renderEventHand === "function") {
      window.renderEventHand(game, els.eventHand);
    }
    if (typeof window.renderStrategyMap === "function") {
      window.renderStrategyMap(game, els.strategyMap);
    }
    if (typeof window.renderSeasonEvents === "function") {
      window.renderSeasonEvents(game, els.events, helpers);
    }
    if (typeof window.renderGrandStrategy === "function") {
      window.renderGrandStrategy(game, { strategy: els.strategy, relations: els.relations });
    }
    if (typeof window.renderDiplomacyIntrigue === "function") {
      window.renderDiplomacyIntrigue(game, { foreign: els.foreign, plots: els.plots }, helpers);
    }
    if (typeof window.renderJusticePanels === "function") {
      window.renderJusticePanels(game, { opinion: els.opinion, cases: els.cases });
    }
    if (typeof window.renderSystemPanels === "function") {
      window.renderSystemPanels(game, { harem: els.harem, heirs: els.heirs, offices: els.offices }, helpers);
    }
    if (typeof window.renderTalentPool === "function") {
      window.renderTalentPool(game, document.querySelector("#talent-list"), helpers);
    }
  }

  window.renderExternalPanels = renderExternalPanels;
})();
