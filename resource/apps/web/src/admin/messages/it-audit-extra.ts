export default {
  messages: {
    userUpdated: "Utente {userId} aggiornato: {changedFields}",
    contentReviewed:
      "Moderazione {type} #{subjectId}: {action}, versione {version}: {reason}",
    topicStatusChanged: 'Stato del topic "{title}" cambiato in {status}',
    topicPinWeightChanged:
      'Peso di fissaggio del topic "{title}" {oldPinWeight} -> {pinWeight}',
    topicCategoriesChanged:
      'Categorie del topic "{title}" {oldCategoryIds} -> {categoryIds}',
    topicDeleted: 'Topic "{title}" eliminato',
    moderatorTopicStatusChanged:
      'Il moderatore ha cambiato il topic "{title}" in {status}',
    categoryModeratorAdded:
      'Aggiunto il moderatore {username} alla categoria "{categoryName}"',
    categoryModeratorRemoved:
      'Rimosso il moderatore {userId} dalla categoria "{categoryName}"',
  },
  statusLabels: {
    blocked: "bloccato",
    unblocked: "normale",
  },
  contentTypes: {
    topic: "topic",
    post: "risposta",
  },
  actions: {
    approve: "approvazione",
    reject: "rifiuto",
    recheck: "nuovo controllo",
  },
  fieldLabels: {
    status: "stato account",
    activation: "stato attivazione",
    role: "ruolo",
  },
} as const;
